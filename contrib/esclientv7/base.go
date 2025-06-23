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

	var result struct {
		Source json.RawMessage `json:"_source"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode response failed: %w", err)
	}

	if err := json.Unmarshal(result.Source, dest); err != nil {
		return fmt.Errorf("unmarshal _source failed: %w", err)
	}

	return nil
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
