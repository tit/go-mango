package mango

import (
  "reflect"
  "testing"
)

func TestClient_stringToSlice(t *testing.T) {
  type fields struct {
    VpbxApiKey  string
    VpbxApiSalt string
  }
  type args struct {
    string string
  }
  tests := []struct {
    name      string
    fields    fields
    args      args
    wantSlice []string
    wantErr   bool
  }{
    {
      name:      `positive`,
      fields:    fields{VpbxApiKey: "", VpbxApiSalt: ""},
      args:      args{string: `[1,2,3]`},
      wantSlice: []string{`1`, `2`, `3`},
      wantErr:   false,
    },
  }
  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      client := &Client{
        VpbxApiKey:  tt.fields.VpbxApiKey,
        VpbxApiSalt: tt.fields.VpbxApiSalt,
      }
      gotSlice, err := client.stringToSlice(tt.args.string)
      if (err != nil) != tt.wantErr {
        t.Errorf("Client.stringToSlice() error = %v, wantErr %v", err, tt.wantErr)
        return
      }
      if !reflect.DeepEqual(gotSlice, tt.wantSlice) {
        t.Errorf("Client.stringToSlice() = %v, want %v", gotSlice, tt.wantSlice)
      }
    })
  }
}

func TestClient_sign(t *testing.T) {
  type fields struct {
    VpbxApiKey  string
    VpbxApiSalt string
  }
  type args struct {
    json string
  }
  tests := []struct {
    name         string
    fields       fields
    args         args
    wantChecksum string
  }{
    {
      name:         `positive`,
      fields:       fields{VpbxApiKey: "", VpbxApiSalt: ""},
      args:         args{json: `{}`},
      wantChecksum: `44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a`,
    },
  }
  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      client := &Client{
        VpbxApiKey:  tt.fields.VpbxApiKey,
        VpbxApiSalt: tt.fields.VpbxApiSalt,
      }
      if gotChecksum := client.sign(tt.args.json); gotChecksum != tt.wantChecksum {
        t.Errorf("Client.sign() = %v, want %v", gotChecksum, tt.wantChecksum)
      }
    })
  }
}
