package cmd

import (
	"testing"
)

func TestExecute(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Execute",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Execute(); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_initConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "initConfig",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initConfig()
		})
	}
}
