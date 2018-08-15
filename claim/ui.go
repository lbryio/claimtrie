package claim

import (
	"bytes"
	"fmt"
	"html/template"
)

func export(n *Node) interface{} {
	hash := ""
	if n.Hash() != nil {
		hash = n.Hash().String()
	}
	return &struct {
		Height     Height
		Hash       string
		Tookover   Height
		NextUpdate Height
		BestClaim  *Claim
		Claims     List
		Supports   List
	}{
		Height:     n.height,
		Hash:       hash,
		Tookover:   n.tookover,
		NextUpdate: n.NextUpdate(),
		BestClaim:  n.best,
		Claims:     n.claims,
		Supports:   n.supports,
	}
}

func nodeToString(n *Node) string {
	ui := ` Height {{.Height}}, {{.Hash}} Tookover: {{.Tookover}} Next: {{.NextUpdate}}
  {{$best := .BestClaim}}
{{- if .Claims}}
  {{range .Claims -}}
  C {{.}} {{if (CMP . $best)}} <B> {{end}}
  {{end}}
{{- end}}
{{- if .Supports}}
  {{range .Supports}}S {{.}}
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

func claimToString(c *Claim) string {
	return fmt.Sprintf("%-68s id: %s accepted: %6d  active: %6d, amt: %12d  effamt: %12d",
		c.OutPoint, c.ID, c.Accepted, c.ActiveAt, c.Amt, c.EffAmt)
}
