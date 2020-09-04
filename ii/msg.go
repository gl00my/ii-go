package ii

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Tags struct {
	Hash map[string]string
	List []string
}

type Msg struct {
	MsgId string
	Tags  Tags
	Echo  string
	Date  int64
	From  string
	Addr  string
	To    string
	Subj  string
	Text  string
}

func MsgId(msg string) string {
	h := sha256.Sum256([]byte(msg))
	id := base64.StdEncoding.EncodeToString(h[:])
	id = strings.Replace(id, "+", "A", -1)
	id = strings.Replace(id, "/", "Z", -1)
	return id[0:20]
}

func IsMsgId(id string) bool {
	return len(id) == 20 && !strings.Contains(id, ".")
}

func IsEcho(e string) bool {
	l := len(e)
	return l >= 3 && l <= 120 && strings.Contains(e, ".")
}

func DecodeMsgline(msg string, enc bool) (*Msg, error) {
	var m Msg
	var data []byte
	var err error
	if enc {
		if data, err = base64.StdEncoding.DecodeString(msg); err != nil {
			if data, err = base64.URLEncoding.DecodeString(msg); err != nil {
				return nil, err
			}
		}
	} else {
		data = []byte(msg)
	}
	text := strings.Split(string(data), "\n")
	if len(text) < 5 {
		return nil, errors.New("Wrong message format")
	}
	if text[3] != "" {
		return nil, errors.New("No body delimiter in message")
	}
	m.Echo = text[0]
	if !IsEcho(m.Echo) {
		return nil, errors.New("Wrong echoarea format")
	}
	m.To = text[1]
	m.Subj = text[2]
	m.Date = time.Now().Unix()
	start := 4
	repto := text[4]
	m.Tags, _ = MakeTags("ii/ok")
	if strings.HasPrefix(repto, "@repto:") {
		start += 1
		repto = strings.Trim(strings.Split(repto, ":")[1], " ")
		m.Tags.Add("repto/" + repto)
		Trace.Printf("Add repto tag: %s", repto)
	}
	for i := start; i < len(text); i++ {
		m.Text += text[i] + "\n"
	}
	m.Text = strings.TrimSuffix(m.Text, "\n")
	Trace.Printf("Final message: %s\n", m.String())
	return &m, nil
}

func DecodeBundle(msg string) (*Msg, error) {
	var m Msg
	if strings.Contains(msg, ":") {
		spl := strings.Split(msg, ":")
		if len(spl) != 2 {
			return nil, errors.New("Wrong bundle format")
		}
		msg = spl[1]
		m.MsgId = spl[0]
		if !IsMsgId(m.MsgId) {
			return nil, errors.New("Wrong MsgId format")
		}
	}
	data, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		return nil, err
	}
	if m.MsgId == "" {
		m.MsgId = MsgId(string(data))
	}
	text := strings.Split(string(data), "\n")
	if len(text) <= 8 {
		return nil, errors.New("Wrong message format")
	}
	m.Tags, err = MakeTags(text[0])
	if err != nil {
		return nil, err
	}
	m.Echo = text[1]
	if !IsEcho(m.Echo) {
		return nil, errors.New("Wrong echoarea format")
	}
	_, err = fmt.Sscanf(text[2], "%d", &m.Date)
	if err != nil {
		return nil, err
	}
	m.From = text[3]
	m.Addr = text[4]
	m.To = text[5]
	m.Subj = text[6]
	for i := 8; i < len(text); i++ {
		m.Text += text[i] + "\n"
	}
	m.Text = strings.TrimSuffix(m.Text, "\n")
	return &m, nil
}

func MakeTags(str string) (Tags, error) {
	var t Tags
	str = strings.Trim(str, " ")
	if str == "" { // empty
		return t, nil
	}
	tags := strings.Split(str, "/")
	if len(tags)%2 != 0 {
		return t, errors.New("Wrong tags: " + str)
	}
	t.Hash = make(map[string]string)
	for i := 0; i < len(tags); i += 2 {
		t.Hash[tags[i]] = tags[i+1]
		t.List = append(t.List, tags[i])
	}
	return t, nil
}

func NewTags(str string) Tags {
	t, _ := MakeTags(str)
	return t
}

func (t *Tags) Add(str string) error {
	tags := strings.Split(str, "/")
	if len(tags)%2 != 0 {
		return errors.New("Wrong tags")
	}
	for i := 0; i < len(tags); i += 2 {
		t.Hash[tags[i]] = tags[i+1]
		t.List = append(t.List, tags[i])
	}
	return nil
}

func (t Tags) String() string {
	var text string
	if t.Hash == nil {
		return ""
	}
	for _, n := range t.List {
		if val, ok := t.Hash[n]; ok {
			text += fmt.Sprintf("%s/%s/", n, val)
		}
	}
	text = strings.TrimSuffix(text, "/")
	return text
}

func (m *Msg) Dump() string {
	if m == nil {
		return ""
	}
	return fmt.Sprintf("id: %s\ntags: %s\nechoarea: %s\ndate: %s\nmsgfrom: %s\naddr: %s\nmsgto: %s\nsubj: %s\n\n%s",
		m.MsgId, m.Tags.String(), m.Echo, time.Unix(m.Date, 0), m.From, m.Addr, m.To, m.Subj, m.Text)
}

func (m *Msg) Tag(n string) (string, bool) {
	if m == nil || m.Tags.Hash == nil {
		return "", false
	}
	v, ok := m.Tags.Hash[n]
	if ok {
		return v, true
	}
	return "", false
}

func (m *Msg) String() string {
	tags := m.Tags.String()
	text := strings.Join([]string{tags, m.Echo,
		fmt.Sprint(m.Date),
		m.From,
		m.Addr,
		m.To,
		m.Subj,
		"",
		m.Text}, "\n")
	return text
}

func (m *Msg) Encode() string {
	var text string
	if m == nil || m.Echo == "" {
		return ""
	}
	if m.Date == 0 {
		now := time.Now()
		m.Date = now.Unix()
	}
	text = m.String()
	if m.MsgId == "" {
		m.MsgId = MsgId(text)
	}
	return m.MsgId + ":" + base64.StdEncoding.EncodeToString([]byte(text))
}
