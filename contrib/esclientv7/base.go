package esclientv7

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"io"
	"strings"
)

type Document struct {
	ID      string
	Content any
}

type Total struct {
	Value    uint64 `json:"value"`
	Relation string `json:"relation"`
}

func (m *Manager) GetByID(ctx context.Context, index, id string, dest any) error {
	req := esapi.GetRequest{
		Index:      index,
		DocumentID: id,
	}
	res, errR := req.Do(ctx, m.client)
	if errR != nil {
		return fmt.Errorf("request error: %w", errR)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return fmt.Errorf("document not found: index=%s id=%s", index, id)
		}
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("es error: %s", string(body))
	}

	if err := json.NewDecoder(res.Body).Decode(dest); err != nil {
		return fmt.Errorf("decode response failed: %w", err)
	}

	return nil
}

type DSLParams struct {
	From  int                 `json:"from"`
	Size  int                 `json:"size,omitempty"`
	Query DSLParamsQuery      `json:"query"`
	Sort  []map[string]string `json:"sort,omitempty"`
}

type DSLParamsQuery struct {
	Bool DSLParamsQueryBool `json:"bool"`
}

type DSLParamsQueryBool struct {
	Must               []any `json:"must,omitempty"`
	Filter             []any `json:"filter,omitempty"`
	MustNot            []any `json:"must_not,omitempty"`
	Should             []any `json:"should,omitempty"`
	MinimumShouldMatch uint8 `json:"minimum_should_match,omitempty"`
}

func (m *Manager) GetDefaultDSLParams() DSLParams {
	return DSLParams{
		Query: DSLParamsQuery{
			Bool: DSLParamsQueryBool{
				Must:               make([]any, 0, 2),
				Filter:             make([]any, 0, 2),
				MustNot:            make([]any, 0, 2),
				Should:             make([]any, 0, 2),
				MinimumShouldMatch: 0,
			},
		},
		Sort: make([]map[string]string, 0, 2),
	}
}

func (m *Manager) DefaultSearch(ctx context.Context, index string, params DSLParams, dest any, o ...func(*esapi.SearchRequest)) (total Total, maxScore float64, err error) {
	return m.Search(ctx, index, params, dest, o...)
}

func (m *Manager) Search(ctx context.Context, index string, dsl any, dest any, o ...func(*esapi.SearchRequest)) (total Total, maxScore float64, err error) {
	var buf bytes.Buffer
	if err = json.NewEncoder(&buf).Encode(dsl); err != nil {
		return
	}

	searchParams := make([]func(*esapi.SearchRequest), 0, 3)
	searchParams = append(searchParams,
		m.client.Search.WithContext(ctx),
		m.client.Search.WithIndex(index),
		m.client.Search.WithBody(&buf),
		//m.client.Search.WithTrackTotalHits(true),
		//m.client.Search.WithPretty(),
	)
	if len(o) > 0 {
		searchParams = append(searchParams, o...)
	}

	var res *esapi.Response
	res, err = m.client.Search(searchParams...)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		err = fmt.Errorf("search error: %s", res.String())
		return
	}

	// 解析结果
	var result struct {
		Hits struct {
			Total    Total           `json:"total"`
			MaxScore float64         `json:"max_score"`
			Hits     json.RawMessage `json:"hits"`
		} `json:"hits"`
	}
	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		err = fmt.Errorf("decode response failed: %w", err)
		return
	}
	total = result.Hits.Total
	maxScore = result.Hits.MaxScore

	if err = json.Unmarshal(result.Hits.Hits, dest); err != nil {
		err = fmt.Errorf("unmarshal hits failed: %w", err)
	}

	return
}

func (m *Manager) InsertDocument(ctx context.Context, index string, doc Document) error {
	data, errJ := json.Marshal(doc.Content)
	if errJ != nil {
		return errJ
	}

	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: doc.ID,
		Body:       bytes.NewReader(data),
		//Refresh:    "true", // 生产中可移除，批量写入时不建议开启
	}

	res, errR := req.Do(ctx, m.client)
	if errR != nil {
		return errR
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("index error: %s", body)
	}
	return nil
}

func (m *Manager) BulkUpsert(ctx context.Context, index string, docs []Document) error {
	var buf bytes.Buffer

	for _, doc := range docs {
		// 第一行：update 指令
		meta := map[string]any{
			"update": map[string]any{
				"_index": index,
				"_id":    doc.ID,
			},
		}
		metaLine, errM := json.Marshal(meta)
		if errM != nil {
			return fmt.Errorf("failed to marshal meta: %w", errM)
		}
		buf.Write(metaLine)
		buf.WriteByte('\n')

		// 第二行：doc + doc_as_upsert
		upsert := map[string]any{
			"doc":           doc.Content,
			"doc_as_upsert": true,
		}
		dataLine, errD := json.Marshal(upsert)
		if errD != nil {
			return fmt.Errorf("failed to marshal data: %w", errD)
		}
		buf.Write(dataLine)
		buf.WriteByte('\n')
	}

	// 发送 Bulk 请求
	res, errB := m.client.Bulk(bytes.NewReader(buf.Bytes()), m.client.Bulk.WithContext(ctx))
	if errB != nil {
		return fmt.Errorf("bulk request failed: %w", errB)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk response error: %s", res.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return fmt.Errorf("failed to decode bulk response: %w", err)
	}

	// 检查是否有失败项
	if resp["errors"].(bool) {
		var failedItems []string
		for _, item := range resp["items"].([]interface{}) {
			for action, result := range item.(map[string]interface{}) {
				status := int(result.(map[string]interface{})["status"].(float64))
				if status >= 300 {
					id := result.(map[string]interface{})["_id"]
					errorReason := result.(map[string]interface{})["error"]
					failedItems = append(failedItems, fmt.Sprintf("ID: %v, Action: %s, Error: %v", id, action, errorReason))
				}
			}
		}
		return fmt.Errorf("bulk upsert partially failed: %s", strings.Join(failedItems, "; "))
	}

	return nil
}
