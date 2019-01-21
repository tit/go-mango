package mango

import (
  "bytes"
  "crypto/sha256"
  "encoding/csv"
  "encoding/json"
  "fmt"
  "io"
  "io/ioutil"
  "net/http"
  "net/url"
  "strconv"
  "strings"
  "time"
)

const (
  apiUrl           string = `https://app.mango-office.ru/vpbx`
  pathUsers               = apiUrl + `/config/users/request`
  pathStatsRequest        = apiUrl + `/stats/request`
  pathStatsResult         = apiUrl + `/stats/result`
  statsFields             = `records,start,finish,answer,from_extension,from_number,to_extension,to_number,disconnect_reason,line_number,location,entry_id`
)

const (
  statsFieldRecords = iota
  statsFieldStart
  statsFieldFinish
  statsFieldAnswer
  statsFieldFromExtension
  statsFieldFromNumber
  statsFieldToExtension
  statsFieldToNumber
  statsFieldDisconnectReason
  statsFieldLineNumber
  statsFieldLocation
  statsFieldEntryId
)

// public structures

// mango class for mango client
// get VpbxApiKey and VpbxApiSalt
// on https://lk.mango-office.ru/api-vpbx/settings
type Client struct {
  VpbxApiKey  string `json:"vpbxApiKey"`
  VpbxApiSalt string `json:"vpbxApiSalt"`
}

// manager, not caller
type User struct {
  General struct {
    Name       string `json:"name"`
    Email      string `json:"email"`
    Department string `json:"department"`
    Position   string `json:"position"`
  } `json:"general"`
  Telephony struct {
    Extension    string `json:"extension"`
    Outgoingline string `json:"outgoingline"`
    Numbers      []struct {
      Number   string `json:"number"`
      Protocol string `json:"protocol"`
      Order    int    `json:"order"`
      WaitSec  int    `json:"wait_sec"`
      Status   string `json:"status"`
    } `json:"numbers"`
  } `json:"telephony"`
}

type Call struct {
  Records          []string
  Start            int64
  Finish           int64
  Answer           int64
  FromExtension    string
  FromNumber       string
  ToExtension      string
  ToNumber         string
  DisconnectReason int
  LineNumber       string
  EntryId          string
  Location         string
}

// private structures

type statsKey struct {
  Key string `json:"key"`
}

type stats struct {
  FromDate  int64  `json:"date_from"`
  ToDate    int64  `json:"date_to"`
  Fields    string `json:"fields"`
  RequestId string `json:"request_id"`
}

func newStats(fromDate int64, toDate int64, requestId string) *stats {
  return &stats{
    FromDate:  fromDate,
    ToDate:    toDate,
    Fields:    statsFields,
    RequestId: requestId,
  }
}

type userExtension struct {
  Extension string `json:"extension"`
}

// public methods

func (client *Client) StatsKey(fromDate time.Time, toDate time.Time, requestId string) (key string, err error) {
  // prepare request data
  fromDateUnixTime := fromDate.Unix()
  toDateUnixTime := toDate.Unix()

  statsRequest := newStats(fromDateUnixTime, toDateUnixTime, requestId)

  // struct to JSON
  jsonBytes, _ := json.Marshal(statsRequest)
  jsonString := string(jsonBytes)

  // JSON to formData
  formData := client.formData(jsonString)

  // request
  response, _ := http.PostForm(pathStatsRequest, formData)
  body, _ := ioutil.ReadAll(response.Body)

  var statsKey statsKey
  _ = json.Unmarshal(body, &statsKey)
  key = statsKey.Key

  return
}

func (client *Client) Stats(key string, requestId string) (calls []Call, err error) {
  requestBody := statsKey{Key: key}
  response, _ := client.post(pathStatsResult, requestBody)
  body, _ := ioutil.ReadAll(response.Body)
  code := response.StatusCode

  switch code {
  case 200:
    calls, _ = client.statsUnmarshal(body)
    return calls, nil
  case 404:
    return calls, fmt.Errorf(`data not found, sended invalid/wrong/expired key`)
  case 204:
    return calls, fmt.Errorf(`data not ready, retry after 5 seconds`)
  default:
    return calls, fmt.Errorf("unknow error, code: %d, body: %s", code, string(body))
  }
}

// Get User By Extension
func (client *Client) User(extension string) (user User, err error) {
  var userExtension userExtension
  userExtension.Extension = extension

  response, err := client.post(pathUsers, userExtension)
  if err != nil {
    return
  }

  body, err := ioutil.ReadAll(response.Body)
  if err != nil {
    return
  }

  var users struct {
    Users []User `json:"users"`
  }

  err = json.Unmarshal(body, &users)
  if err != nil {
    return
  }

  if len(users.Users) == 0 {
    err = fmt.Errorf("user %s not found", extension)
    return
  }

  return users.Users[0], err
}

// private methods

func (client *Client) statsUnmarshal(body []byte) (calls []Call, err error) {
  reader := csv.NewReader(bytes.NewReader(body))
  reader.Comma = ';'

  for {
    line, err := reader.Read()
    if err == io.EOF {
      break
    } else if err != nil {
      return calls, fmt.Errorf(`error read CSV`)
    }

    record, _ := client.stringToSlice(line[statsFieldRecords])
    start, _ := strconv.ParseInt(line[statsFieldStart], 10, 64)
    finish, _ := strconv.ParseInt(line[statsFieldFinish], 10, 64)
    answer, _ := strconv.ParseInt(line[statsFieldAnswer], 10, 64)
    disconnectReason, _ := strconv.Atoi(line[statsFieldDisconnectReason])
    calls = append(calls, Call{
      Records:          record,
      Start:            start,
      Finish:           finish,
      Answer:           answer,
      FromExtension:    line[statsFieldFromExtension],
      FromNumber:       line[statsFieldFromNumber],
      ToExtension:      line[statsFieldToExtension],
      ToNumber:         line[statsFieldToNumber],
      DisconnectReason: disconnectReason,
      LineNumber:       line[statsFieldLineNumber],
      Location:         line[statsFieldLocation],
      EntryId:          line[statsFieldEntryId],
    })
  }

  return
}

func (client *Client) post(url string, v interface{}) (response *http.Response, err error) {
  jsonBytes, _ := json.Marshal(v)
  jsonString := string(jsonBytes)
  formData := client.formData(jsonString)
  response, _ = http.PostForm(url, formData)
  return
}

func (client *Client) formData(json string) (formData url.Values) {
  formData = url.Values{
    "vpbx_api_key": {client.VpbxApiKey},
    "sign":         {client.sign(json)},
    "json":         {json},
  }
  return
}

func (client *Client) sign(json string) (checksum string) {
  sum256 := sha256.Sum256([]byte(client.VpbxApiKey + json + client.VpbxApiSalt))
  checksum = fmt.Sprintf("%x", sum256)

  return
}

func (client *Client) statsFields() (fields []string) {
  fields = []string{
    "records",
    "start",
    "finish",
    "answer",
    "from_extension",
    "from_number",
    "to_extension",
    "to_number",
    "disconnect_reason",
    "line_number",
    "location",
    "entry_id",
  }

  return
}

func (client *Client) stringToSlice(string string) (slice []string, err error) {
  slice = strings.Split(string[1:][:len(string)-2], `,`)

  return
}
