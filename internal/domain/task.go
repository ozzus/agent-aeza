package domain

import "time"

// тип задачи для проверки

type TaskType string

const (
	TaskTypeHTTP       TaskType = "http"
	TaskTypePing       TaskType = "ping"
	TaskTypeTCP        TaskType = "tcp"
	TaskTypeTraceroute TaskType = "traceroute"
	TaskTypeDNS        TaskType = "dns"
)

//типы DNS записей

type DNSRecordType string

const (
	DNSRecordA    DNSRecordType = "A"
	DNSRecordAAAA DNSRecordType = "AAAA"
	DNSRecordMX   DNSRecordType = "MX"
	DNSRecordNS   DNSRecordType = "NS"
	DNSRecordTXT  DNSRecordType = "TXT"
)

type Task struct {
	ID          string                 `json:"task_id"`
	Type        TaskType               `json:"type"`
	Target      string                 `json:"target"`
	Parameters  map[string]interface{} `json:"parameters"`
	ScheduledAt time.Time              `json:"scheduled_at"`
	CreatedAt   time.Time              `json:"created_at"`
	Timeout     int                    `json:"timeout"`
}

// Список задач
type TaskList struct {
	Tasks []Task `json:"tasks"`
	Total int    `json:"total"`
}
