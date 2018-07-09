package claimnode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"sort"

	"github.com/lbryio/claimtrie/claim"
)

func sortedBestClaims(n *Node) []string {
	var s []string
	for i := claim.Height(0); i <= n.Tookover(); i++ {
		v, ok := n.bestClaims[i]
		if !ok || v == nil {
			continue
		}
		s = append(s, fmt.Sprintf("{%d, %d}, ", i, v.OutPoint.Index))
	}
	return s
}

func sortedClaims(n *Node) []*claim.Claim {
	c := make([]*claim.Claim, 0, len(n.claims))
	for _, v := range n.claims {
		c = append(c, v)
	}
	sort.Slice(c, func(i, j int) bool { return c[i].Seq < c[j].Seq })
	return c
}

func sortedSupports(n *Node) []*claim.Support {
	s := make([]*claim.Support, 0, len(n.supports))
	for _, v := range n.supports {
		s = append(s, v)
	}
	sort.Slice(s, func(i, j int) bool { return s[i].Seq < s[j].Seq })
	return s
}

func export(n *Node) interface{} {
	return &struct {
		Height     claim.Height
		Hash       string
		BestClaims []string
		BestClaim  *claim.Claim
		Claims     []*claim.Claim
		Supports   []*claim.Support
	}{
		Height:     n.height,
		Hash:       n.Hash().String(),
		BestClaims: sortedBestClaims(n),
		BestClaim:  n.BestClaim(),
		Claims:     sortedClaims(n),
		Supports:   sortedSupports(n),
	}
}

func toString(n *Node) string {
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

	w := bytes.NewBuffer(nil)
	t := template.Must(template.New("").Funcs(template.FuncMap{
		"CMP": func(a, b *claim.Claim) bool { return a == b },
	}).Parse(ui))
	if err := t.Execute(w, export(n)); err != nil {
		fmt.Printf("can't execute template, err: %s\n", err)
	}
	return w.String()
}

func toJSON(n *Node) ([]byte, error) {
	return json.Marshal(export(n))
}
