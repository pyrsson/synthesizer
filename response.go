package main

var dummyResponse = map[string]any{
	"requestId":   "a1b2c3d4",
	"generatedAt": "2025-11-17T12:34:56Z",
	"metrics": map[string]any{
		"latencyMs":   123,
		"successRate": 0.987,
		"errors": []map[string]any{
			{"code": "TIMEOUT", "count": 3},
			{"code": "BAD_REQUEST", "count": 1},
		},
	},
	"users": []map[string]any{
		{
			"id":     1,
			"name":   "Alice",
			"active": true,
			"roles":  []string{"admin", "tester"},
			"tags": map[string]string{
				"region": "eu-west",
				"plan":   "pro",
			},
		},
		{
			"id":     2,
			"name":   "Bob",
			"active": false,
			"roles":  []string{"viewer"},
			"tags": map[string]string{
				"region": "us-east",
				"plan":   "free",
			},
		},
	},
}
