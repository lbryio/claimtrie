package claim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"sort"
)

func sortedBestClaims(n *Node) []string {
	var s []string
	for i := Height(0); i <= n.Tookover(); i++ {
		v, ok := n.bestClaims[i]
		if !ok || v == nil {
			continue
		}
		s = append(s, fmt.Sprintf("{%d, %d}, ", i, v.OutPoint.Index))
	}
	return s
}

func sortedClaims(n *Node) []*Claim {
	c := make([]*Claim, len(n.claims))
	copy(c, n.claims)
	sort.Slice(c, func(i, j int) bool { return c[i].seq < c[j].seq })
	return c
}

func sortedSupports(n *Node) []*Support {
	s := make([]*Support, len(n.supports))
	copy(s, n.supports)
	sort.Slice(s, func(i, j int) bool { return s[i].seq < s[j].seq })
	return s
}

func export(n *Node) interface{} {
	return &struct {
		Height     Height
		Hash       string
		BestClaims []string
		BestClaim  *Claim
		Claims     []*Claim
		Supports   []*Support
	}{
		Height:     n.height,
		Hash:       n.Hash().String(),
		BestClaims: sortedBestClaims(n),
		BestClaim:  n.BestClaim(),
		Claims:     sortedClaims(n),
		Supports:   sortedSupports(n),
	}
}

func nodeToString(n *Node) string {
	ui := ` Height {{.Height}}, {{.Hash}} BestClaims: {{range .BestClaims}}{{.}}{{end}}
  {{$best := .BestClaim}}
{{- if .Claims}}
  {{range .Claims -}}
  {{.}} {{if (CMP . $best)}} <B> {{end}}
  {{end}}
{{- end}}
{{- if .Supports}}
  {{range .Supports}}{{.}}
  {{end}}
{{- end}}`

	t := template.Must(template.New("").Funcs(template.FuncMap{
		"CMP": func(a, b *Claim) bool { return a == b },
	}).Parse(ui))
	w := bytes.NewBuffer(nil)
	if err := t.Execute(w, export(n)); err != nil {
		fmt.Printf("can't execute template, err: %s\n", err)
	}
	return w.String()
}

func nodeToJSON(n *Node) ([]byte, error) {
	return json.Marshal(export(n))
}

func claimToString(c *Claim) string {
	return fmt.Sprintf("C-%-68s amt: %-3d  effamt: %-3d  accepted: %-3d  active: %-3d  id: %s",
		c.OutPoint, c.Amt, c.EffAmt, c.Accepted, c.ActiveAt, c.ID)
}

func claimToJSON(c *Claim) ([]byte, error) {
	return json.Marshal(&struct {
		OutPoint  string
		ID        string
		Amount    Amount
		EffAmount Amount
		Accepted  Height
		ActiveAt  Height
	}{
		OutPoint:  c.OutPoint.String(),
		ID:        c.ID.String(),
		Amount:    c.Amt,
		EffAmount: c.EffAmt,
		Accepted:  c.Accepted,
		ActiveAt:  c.ActiveAt,
	})
}

func supportToString(s *Support) string {
	return fmt.Sprintf("S-%-68s amt: %-3d               accepted: %-3d  active: %-3d  id: %s",
		s.OutPoint, s.Amt, s.Accepted, s.ActiveAt, s.ClaimID)
}
func supportToJSON(s *Support) ([]byte, error) {
	return json.Marshal(&struct {
		OutPoint string
		ClaimID  string
		Amount   Amount
		Accepted Height
		ActiveAt Height
	}{
		OutPoint: s.OutPoint.String(),
		ClaimID:  s.ClaimID.String(),
		Amount:   s.Amt,
		Accepted: s.Accepted,
		ActiveAt: s.ActiveAt,
	})
}
