package model

import "testing"

func ptr(s string) *string { return &s }

func TestCreateTaskRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		req       CreateTaskRequest
		wantField string // empty means no errors expected
	}{
		{
			name:      "valid minimal",
			req:       CreateTaskRequest{Title: "Do thing"},
			wantField: "",
		},
		{
			name:      "missing title",
			req:       CreateTaskRequest{},
			wantField: "title",
		},
		{
			name:      "invalid status",
			req:       CreateTaskRequest{Title: "X", Status: "banana"},
			wantField: "status",
		},
		{
			name:      "invalid priority",
			req:       CreateTaskRequest{Title: "X", Priority: "urgent"},
			wantField: "priority",
		},
		{
			name:      "invalid assignee_email",
			req:       CreateTaskRequest{Title: "X", AssigneeEmail: ptr("not-an-email")},
			wantField: "assignee_email",
		},
		{
			name:      "valid assignee_email",
			req:       CreateTaskRequest{Title: "X", AssigneeEmail: ptr("user@example.com")},
			wantField: "",
		},
		{
			name:      "empty assignee_email normalizes to nil",
			req:       CreateTaskRequest{Title: "X", AssigneeEmail: ptr("")},
			wantField: "",
		},
		{
			name:      "invalid due_date",
			req:       CreateTaskRequest{Title: "X", DueDate: ptr("not-a-date")},
			wantField: "due_date",
		},
		{
			name:      "valid due_date",
			req:       CreateTaskRequest{Title: "X", DueDate: ptr("2026-05-01")},
			wantField: "",
		},
		{
			name:      "empty due_date normalizes to nil",
			req:       CreateTaskRequest{Title: "X", DueDate: ptr("")},
			wantField: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.req.Validate()
			if tt.wantField == "" {
				if len(errs) != 0 {
					t.Fatalf("expected no errors, got %v", errs)
				}
			} else {
				if _, ok := errs[tt.wantField]; !ok {
					t.Fatalf("expected error on field %q, got %v", tt.wantField, errs)
				}
			}
		})
	}
}

func TestUpdateTaskRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		req       UpdateTaskRequest
		wantField string
	}{
		{
			name:      "all nil is valid",
			req:       UpdateTaskRequest{},
			wantField: "",
		},
		{
			name:      "empty title",
			req:       UpdateTaskRequest{Title: ptr("")},
			wantField: "title",
		},
		{
			name:      "invalid status",
			req:       UpdateTaskRequest{Status: ptr("yolo")},
			wantField: "status",
		},
		{
			name:      "invalid assignee_email",
			req:       UpdateTaskRequest{AssigneeEmail: ptr("garbage")},
			wantField: "assignee_email",
		},
		{
			name:      "empty assignee_email clears (valid)",
			req:       UpdateTaskRequest{AssigneeEmail: ptr("")},
			wantField: "",
		},
		{
			name:      "invalid due_date",
			req:       UpdateTaskRequest{DueDate: ptr("13/99/2026")},
			wantField: "due_date",
		},
		{
			name:      "empty due_date clears (valid)",
			req:       UpdateTaskRequest{DueDate: ptr("")},
			wantField: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.req.Validate()
			if tt.wantField == "" {
				if len(errs) != 0 {
					t.Fatalf("expected no errors, got %v", errs)
				}
			} else {
				if _, ok := errs[tt.wantField]; !ok {
					t.Fatalf("expected error on field %q, got %v", tt.wantField, errs)
				}
			}
		})
	}
}
