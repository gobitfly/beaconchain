package secrets

import (
	"context"
	"fmt"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

var ctx = context.Background()
var client, _ = secretmanager.NewClient(ctx)

func fetchSecret(secretName string) (string, error) {
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName,
	}
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %w", err)
	}
	return string(result.Payload.Data), nil
}

// "changedKeys" is output param of replaced keys; nested keys are formatted as "parent.child"
func ReplaceNestedSecrets(parentKey string, data map[string]interface{}, changedKeys map[string]string) error {
	for k, v := range data {
		fullKey := k
		if parentKey != "" {
			fullKey = fmt.Sprintf("%s.%s", parentKey, k)
		}
		switch val := v.(type) {
		case string:
			if strings.HasPrefix(val, "gsm:") {
				secretVal, err := fetchSecret(strings.TrimPrefix(val, "gsm:"))
				if err != nil {
					return err
				}
				changedKeys[fullKey] = secretVal
			}
		case map[string]interface{}:
			if err := ReplaceNestedSecrets(fullKey, val, changedKeys); err != nil {
				return err
			}
		}
	}
	return nil
}
