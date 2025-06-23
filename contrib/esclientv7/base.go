package esclientv7

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"io"
)

func (m *Manager) InsertDocument(ctx context.Context, index string, docID string, doc interface{}) error {
	data, errJ := json.Marshal(doc)
	if errJ != nil {
		return errJ
	}

	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: docID,
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
