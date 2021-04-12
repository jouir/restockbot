package main

import (
	"fmt"
	"testing"
)

func TestExtractShopName(t *testing.T) {
	tests := []struct {
		link string // url to parse
		name string // expected name
	}{
		{"https://www.topachat.com/pages/produits_cat_est_micro_puis_rubrique_est_wgfx_pcie_puis_f_est_58-11733,11575,11447,11445,11446,10587,11796,11559,11558,11586.html", "topachat.com"},
		{"https://www.ldlc.com/informatique/pieces-informatique/carte-graphique-interne/c4684/+fv121-17715,19183,19184,19185,19339,19340,19365,19367,19509,19674.html", "ldlc.com"},
		{"https://www.cybertek.fr/carte-graphique-6.aspx?crits=3991%3a4236%3a4387%3a4237%3a4242%3a4289%3a4229%3a4144%3a4145%3a4146", "cybertek.fr"},
		{"https://www.mediamarkt.ch/fr/category/_cartes-graphiques-751073.html", "mediamarkt.ch"},
		{"https://www.steg-electronics.ch/fr/product/pool/nvidia-3060", "steg-electronics.ch"},
		{"https://www.vsgamers.es/category/componentes/tarjetas-graficas?filter-modelo=rtxr-3060-1307", "vsgamers.es"},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("TestExtractShopName#%d", i), func(t *testing.T) {
			name, err := ExtractShopName(tc.link)
			if err != nil {
				t.Errorf("for %s: got %s, want %s", tc.link, err, tc.name)
			} else if name != tc.name {
				t.Errorf("for %s: got %s, want %s", tc.link, name, tc.name)
			} else {
				t.Logf("for %s: got %s, want %s", tc.link, name, tc.name)
			}
		})
	}
}
