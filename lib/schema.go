package gorncs

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"
)

// XMLBilans structure Bilans XML
type XMLBilans struct {
	XMLName xml.Name `xml:"bilans"`
	Bilan   XMLBilan `xml:"bilan"`
}

// XMLBilan structure Bilan XML
type XMLBilan struct {
	XMLName  xml.Name    `xml:"bilan"`
	Identite XMLIdentite `xml:"identite"`
	Detail   XMLDetail   `xml:"detail"`
}

// XMLIdentite structure Identite XML
type XMLIdentite struct {
	XMLName                      xml.Name `xml:"identite"`
	Siren                        string   `xml:"siren"`
	DateClotureExercice          string   `xml:"date_cloture_exercice"`
	CodeGreffe                   string   `xml:"code_greffe"`
	NumDepot                     string   `xml:"num_depot"`
	NumGestion                   string   `xml:"num_gestion"`
	CodeActivite                 string   `xml:"code_activite"`
	DateClotureExercicePrecedent string   `xml:"date_cloture_exercice_n-1"`
	DureeExercice                string   `xml:"duree_exercice_n"`
	DureeExercicePrecedent       string   `xml:"duree_exercice_n-1"`
	DateDepot                    string   `xml:"date_depot"`
	CodeMotif                    string   `xml:"code_motif"`
	CodeTypeBilan                string   `xml:"code_type_bilan"`
	CodeDevise                   string   `xml:"code_devise"`
	CodeOrigineDevise            string   `xml:"code_origine_devise"`
	CodeConfidentialite          string   `xml:"code_confidentialite"`
	Denomination                 string   `xml:"denomination"`
	Adresse                      string   `xml:"adresse"`
}

// XMLDetail structure Detail XML
type XMLDetail struct {
	XMLName xml.Name  `xml:"detail"`
	Page    []XMLPage `xml:"page"`
}

// XMLPage structure Page XML
type XMLPage struct {
	XMLName xml.Name    `xml:"page"`
	Numero  string      `xml:"numero,attr"`
	Liasse  []XMLLiasse `xml:"liasse"`
}

// XMLLiasse structure Liasse XML
type XMLLiasse struct {
	XMLName xml.Name `xml:"liasse"`
	Code    string   `xml:"code,attr"`
	M1      *int     `xml:"m1,string,attr"`
	M2      *int     `xml:"m2,string,attr"`
	M3      *int     `xml:"m3,string,attr"`
	M4      *int     `xml:"m4,string,attr"`
}

// Key Identifiant d'une ligne de bilan
type Key struct {
	CodeBilan string `json:"codeBilan"`
	CodePoste string `json:"codePoste"`
}

// Values données portées par une ligne
// Chaque ligne comporte 4 colonnes
type Values struct {
	M1 *int `json:"valeur_m1"`
	M2 *int `json:"valeur_m2"`
	M3 *int `json:"valeur_m3"`
	M4 *int `json:"valeur_m4"`
}

// Bilan objet restructurant les données d'un bilan au format XML rncs en format maison.
// Les lignes sont une map identifiées par les postes fournis dans la variable Kb
type Bilan struct {
	Siren                        string          `json:"siren" bson:"siren"`
	DateClotureExercice          time.Time       `json:"dateClotureExercice" bson:"dateClotureExercice"`
	CodeGreffe                   string          `json:"codeGreffe" bson:"codeGreffe"`
	NumDepot                     string          `json:"numDepot" bson:"numDepot"`
	NumGestion                   string          `json:"numGestion" bson:"numGestion"`
	CodeActivite                 string          `json:"codeActivite" bson:"codeActivite"`
	DateClotureExercicePrecedent time.Time       `json:"dateClotureExercicePrecedent" bson:"dateClotureExercicePrecedent"`
	DureeExercice                string          `json:"dureeExercice" bson:"dureeExercice"`
	DureeExercicePrecedent       string          `json:"dureeExercicePrecedent" bson:"dureeExercicePrecedent"`
	DateDepot                    time.Time       `json:"dateDepot" bson:"dateDepot"`
	CodeMotif                    string          `json:"codeMotif" bson:"codeMotif"`
	CodeTypeBilan                string          `json:"codeTypeBilan" bson:"codeTypeBilan"`
	CodeDevise                   string          `json:"codeDevise" bson:"codeDevise"`
	CodeOrigineDevise            string          `json:"codeOrigineDevise" bson:"codeOrigineDevise"`
	CodeConfidentialite          string          `json:"codeConfidentialite" bson:"codeConfidentialite"`
	Denomination                 string          `json:"denomination" bson:"denomination"`
	Adresse                      string          `json:"adresse" bson:"adresse"`
	XMLSource                    string          `json:"XMLSource" bson:"XMLSource"`
	NomFichier                   string          `json:"nomFichier" bson:"nomFichier"`
	Report                       []string        `json:"RapportConversion"`
	Lignes                       map[string]*int `json:"lignes" bson:"lignes"`
}

// Postes liste des postes extrats de la variable Kb
var Postes = getPostes()
var PostesDetail = getPostesDetail()

// getPostes retourne la liste des postes extraits de la variable Kb
func getPostes() []string {
	var dbSchema = make(map[string]struct{})
	for codePoste := range Kb {
		for codeBilan := range Kb[codePoste] {
			key := Key{CodeBilan: codeBilan, CodePoste: codePoste}
			schema, _ := GetSchema(key)
			if schema[0] != "" {
				dbSchema[schema[0]] = struct{}{}
			}
			if schema[1] != "" {
				dbSchema[schema[1]] = struct{}{}
			}
			if schema[2] != "" {
				dbSchema[schema[2]] = struct{}{}
			}
			if schema[3] != "" {
				dbSchema[schema[3]] = struct{}{}
			}
		}
	}
	var postes []string
	for k := range dbSchema {
		postes = append(postes, k)
	}
	sort.Slice(postes, func(a int, b int) bool { return postes[a] < postes[b] })
	return postes
}

// Schema décrit les emplacements où peut se trouver un champ
type Schema struct {
	Page    string `json:"page"`
	Code    string `json:"code"`
	Bilan   string `json:"bilan"`
	Colonne int    `json:"colonne"`
}

var bilans = map[string]string{
	"S": "simplifie",
	"C": "complet",
	"K": "consolide",
	"B": "banque",
	"A": "assurance",
}

// getPostesDetail Produit un catalogue des champs avec les informations de schéma
func getPostesDetail() map[string][]Schema {
	postesDetail := make(map[string][]Schema)
	for codePoste := range Kb {
		for codeBilan := range Kb[codePoste] {
			key := Key{CodeBilan: codeBilan, CodePoste: codePoste}
			schema, _ := GetSchema(key)
			if schema[0] != "" {
				s := Schema{
					Page:    Kb[codePoste][codeBilan][0],
					Code:    codePoste,
					Bilan:   bilans[codeBilan],
					Colonne: 1,
				}
				postesDetail[schema[0]] = append(postesDetail[schema[0]], s)
			}
			if schema[1] != "" {
				s := Schema{
					Page:    Kb[codePoste][codeBilan][0],
					Code:    codePoste,
					Bilan:   bilans[codeBilan],
					Colonne: 2,
				}
				postesDetail[schema[1]] = append(postesDetail[schema[1]], s)
			}
			if schema[2] != "" {
				s := Schema{
					Page:    Kb[codePoste][codeBilan][0],
					Code:    codePoste,
					Bilan:   bilans[codeBilan],
					Colonne: 3,
				}
				postesDetail[schema[2]] = append(postesDetail[schema[2]], s)
			}
			if schema[3] != "" {
				s := Schema{
					Page:    Kb[codePoste][codeBilan][0],
					Code:    codePoste,
					Bilan:   bilans[codeBilan],
					Colonne: 4,
				}
				postesDetail[schema[3]] = append(postesDetail[schema[3]], s)
			}
		}
	}
	return postesDetail
}

// toNullString produit une variable sql.NullString à partir d'une string.
// "" devient NULL.
func toNullString(s string) sql.NullString {
	if s != "" {
		return sql.NullString{
			String: s,
			Valid:  true,
		}
	}
	return sql.NullString{
		Valid: false,
	}
}

// GetCreateTableQuery fournit la requête de création de la table sqlite
func GetCreateTableQuery() string {
	createTableQuery := `create table bilan (
		id integer primary key,
		nom_fichier text,
		siren text,
		date_cloture_exercice datetime,
		code_greffe text,
		num_depot text,
		num_gestion text,
		code_activite text,
		date_cloture_exercice_precedent datetime,
		duree_exercice text,
		duree_exercice_precedent text,
		date_depot datetime,
		code_motif text,
		code_type_bilan text,
		code_devise text,
		code_origine_devise text,
		code_confidentialite text,
		denomination text,
		adresse text,
		rapport_integration text`
	for _, p := range Postes {
		createTableQuery = createTableQuery + `,
		` + p + ` integer`
	}
	createTableQuery = createTableQuery + ");"
	return createTableQuery
}

// GetQueryString computes insert query from Kb
func GetQueryString() string {
	l := len(Postes)

	var params []string
	for i := 1; i < l+20; i++ {
		params = append(params, "$"+strconv.Itoa(i))
	}

	query := `insert into bilan (nom_fichier, siren, date_cloture_exercice, code_greffe, 
		num_depot, num_gestion, code_activite, date_cloture_exercice_precedent,
		duree_exercice, duree_exercice_precedent,	date_depot, code_motif,
		code_type_bilan, code_devise, code_origine_devise, code_confidentialite, 
		denomination, adresse, rapport_integration, ` + strings.Join(Postes, ", ") +
		`) values (` + strings.Join(params, ", ") + `);`

	return query
}

// ToQueryParams transforme un objet Bilan en liste de champs à insérer dans une base de données
func (bilan Bilan) ToQueryParams() []interface{} {
	var params []interface{}
	params = append(params, toNullString(bilan.NomFichier))
	params = append(params, toNullString(bilan.Siren))
	params = append(params, bilan.DateClotureExercice)
	params = append(params, toNullString(bilan.CodeGreffe))
	params = append(params, toNullString(bilan.NumDepot))
	params = append(params, toNullString(bilan.NumGestion))
	params = append(params, toNullString(bilan.CodeActivite))
	params = append(params, bilan.DateClotureExercicePrecedent)
	params = append(params, toNullString(bilan.DureeExercice))
	params = append(params, toNullString(bilan.DureeExercicePrecedent))
	params = append(params, bilan.DateDepot)
	params = append(params, toNullString(bilan.CodeMotif))
	params = append(params, toNullString(bilan.CodeTypeBilan))
	params = append(params, toNullString(bilan.CodeDevise))
	params = append(params, toNullString(bilan.CodeOrigineDevise))
	params = append(params, toNullString(bilan.CodeConfidentialite))
	params = append(params, toNullString(bilan.Denomination))
	params = append(params, toNullString(bilan.Adresse))
	rapportIntegration, _ := json.Marshal(bilan.Report)
	if string(rapportIntegration) == "null" {
		params = append(params, toNullString(""))
	} else {
		params = append(params, toNullString(string(rapportIntegration)))
	}

	for _, p := range Postes {
		if l, ok := bilan.Lignes[p]; ok {
			params = append(params, *l)
		} else {
			params = append(params, nil)
		}
	}

	return params
}

// GetSchema convertir une clé en schéma
// un schéma comporte l'intitulé de colonne pour les 4 colonnes présentes dans le fichier source
func GetSchema(key Key) ([4]string, error) {
	if poste, ok := Kb[key.CodePoste]; ok {
		if s, ok := poste[key.CodeBilan]; ok {
			var schema [4]string
			if s[2] != "" {
				schema[0] = s[1] + "_" + s[2]
			}
			if s[3] != "" {
				schema[1] = s[1] + "_" + s[3]
			}
			if s[4] != "" {
				schema[2] = s[1] + "_" + s[4]
			}
			if s[5] != "" {
				schema[3] = s[1] + "_" + s[5]
			}
			return schema, nil
		}
		// certains bilans ont un type complet mais des lignes issues du type simplifié
		// ci-dessous workaround pour prendre en compte ces lignes
		s := poste["S"]
		var schema [4]string
		if s[2] != "" {
			schema[0] = s[1] + "_" + s[2]
		}
		if s[3] != "" {
			schema[1] = s[1] + "_" + s[3]
		}
		if s[4] != "" {
			schema[2] = s[1] + "_" + s[4]
		}
		if s[5] != "" {
			schema[3] = s[1] + "_" + s[5]
		}
		return schema, nil
	}

	return [4]string{"", "", "", ""}, errors.New("schema introuvable pour le type de bilan")
}

// KB type qui contient le schema cible
// **structure**
// { "code_poste": {
//	   "code_bilan": {
//	     "page",
//       "rubrique",
//       "intitule_colonne_1",
//       "intitule_colonne_2",
//       "intitule_colonne_3",
//       "intitule_colonne_4",
//     }
//   }
// }
type KB map[string]map[string][6]string

// Kb schema cible. Les postes du bilan sont calculés à partir de cette variable.
// le champ M3 d'une ligne AA d'un bilan de type complet deviendra actif_capital_souscrit_non_appele_net
var Kb = KB{
	"AA": {
		"C": {
			"01",
			"actif",
			"capital_souscrit_non_appele_brut",
			"",
			"capital_souscrit_non_appele_net",
			"capital_souscrit_non_appele_net_n1",
		},
		"K": {
			"01",
			"actif",
			"capital_souscrit_non_appele_brut",
			"",
			"capital_souscrit_non_appele_net",
			"capital_souscrit_non_appele_net_n1",
		},
	},
	"AB": {
		"C": {
			"01",
			"actif",
			"frais_etablissement_brut",
			"frais_etablissement_amortissement",
			"frais_etablissement_net",
			"frais_etablissement_net_n1",
		},
		"K": {
			"01",
			"actif",
			"frais_etablissement_brut",
			"frais_etablissement_amortissement",
			"frais_etablissement_net",
			"frais_etablissement_net_n1",
		},
	},
	"CX": {
		"C": {
			"01",
			"actif",
			"frais_de_developpement_ou_de_recherche_et_developpement_brut",
			"frais_de_developpement_ou_de_recherche_et_developpement_amortissement",
			"frais_de_developpement_ou_de_recherche_et_developpement_net",
			"frais_de_developpement_ou_de_recherche_et_developpement_net_n1",
		},
		"K": {
			"01",
			"actif",
			"frais_de_developpement_ou_de_recherche_et_developpement_brut",
			"frais_de_developpement_ou_de_recherche_et_developpement_amortissement",
			"frais_de_developpement_ou_de_recherche_et_developpement_net",
			"frais_de_developpement_ou_de_recherche_et_developpement_net_n1",
		},
	},
	"AF": {
		"C": {
			"01",
			"actif",
			"concessions_brevets_droits_similaires_brut",
			"concessions_brevets_droits_similaires_amortissement",
			"concessions_brevets_droits_similaires_net",
			"concessions_brevets_droits_similaires_net_n1",
		},
		"K": {
			"01",
			"actif",
			"concessions_brevets_droits_similaires_brut",
			"concessions_brevets_droits_similaires_amortissement",
			"concessions_brevets_droits_similaires_net",
			"concessions_brevets_droits_similaires_net_n1",
		},
	},
	"AH": {
		"C": {
			"01",
			"actif",
			"fond_commercial_brut",
			"fond_commercial_amortissement",
			"fond_commercial_net",
			"fond_commercial_net_n1",
		},
		"K": {
			"01",
			"actif",
			"fond_commercial_brut",
			"fond_commercial_amortissement",
			"fond_commercial_net",
			"fond_commercial_net_n1",
		},
	},
	"AJ": {
		"C": {
			"01",
			"actif",
			"autres_immobilisations_incorporelles_brut",
			"autres_immobilisations_incorporelles_amortissement",
			"autres_immobilisations_incorporelles_net",
			"autres_immobilisations_incorporelles_net_n1",
		},
		"K": {
			"01",
			"actif",
			"autres_immobilisations_incorporelles_brut",
			"autres_immobilisations_incorporelles_amortissement",
			"autres_immobilisations_incorporelles_net",
			"autres_immobilisations_incorporelles_net_n1",
		},
	},
	"AL": {
		"C": {
			"01",
			"actif",
			"avances_et_acomptes_sur_immobilisations_incorporelles_brut",
			"avances_et_acomptes_sur_immobilisations_incorporelles_amortissement",
			"avances_et_acomptes_sur_immobilisations_incorporelles_net",
			"avances_et_acomptes_sur_immobilisations_incorporelles_net_n1",
		},
		"K": {
			"01",
			"actif",
			"avances_et_acomptes_sur_immobilisations_incorporelles_brut",
			"avances_et_acomptes_sur_immobilisations_incorporelles_amortissement",
			"avances_et_acomptes_sur_immobilisations_incorporelles_net",
			"avances_et_acomptes_sur_immobilisations_incorporelles_net_n1",
		},
	},
	"AN": {
		"C": {
			"01",
			"actif",
			"terrains_brut",
			"terrains_amortissement",
			"terrains_net",
			"terrains_net_n1",
		},
		"K": {
			"01",
			"actif",
			"terrains_brut",
			"terrains_amortissement",
			"terrains_net",
			"terrains_net_n1",
		},
	},
	"AR": {
		"C": {
			"01",
			"actif",
			"installations_techniques_materiel_et_outillage_indutriels_brut",
			"installations_techniques_materiel_et_outillage_indutriels_amortissement",
			"installations_techniques_materiel_et_outillage_indutriels_net",
			"installations_techniques_materiel_et_outillage_indutriels_net_n1",
		},
		"K": {
			"01",
			"actif",
			"installations_techniques_materiel_et_outillage_indutriels_brut",
			"installations_techniques_materiel_et_outillage_indutriels_amortissement",
			"installations_techniques_materiel_et_outillage_indutriels_net",
			"installations_techniques_materiel_et_outillage_indutriels_net_n1",
		},
	},
	"AP": {
		"C": {
			"01",
			"actif",
			"constructions_brut",
			"constructions_amortissement",
			"constructions_net",
			"constructions_net_n1",
		},
		"K": {
			"01",
			"actif",
			"constructions_brut",
			"constructions_amortissement",
			"constructions_net",
			"constructions_net_n1",
		},
	},
	"AT": {
		"C": {
			"01",
			"actif",
			"autres_immobilisations_corporelles_brut",
			"autres_immobilisations_corporelles_amortissement",
			"autres_immobilisations_corporelles_net",
			"autres_immobilisations_corporelles_net_n1",
		},
		"K": {
			"01",
			"actif",
			"autres_immobilisations_corporelles_brut",
			"autres_immobilisations_corporelles_amortissement",
			"autres_immobilisations_corporelles_net",
			"autres_immobilisations_corporelles_net_n1",
		},
	},
	"AV": {
		"C": {
			"01",
			"actif",
			"immobilisations_en_cours_brut",
			"immobilisations_en_cours_amortissement",
			"immobilisations_en_cours_net",
			"immobilisations_en_cours_net_n1",
		},
		"K": {
			"01",
			"actif",
			"immobilisations_en_cours_brut",
			"immobilisations_en_cours_amortissement",
			"immobilisations_en_cours_net",
			"immobilisations_en_cours_net_n1",
		},
	},
	"AX": {
		"C": {
			"01",
			"actif",
			"avances_et_acomptes_brut",
			"avances_et_acomptes_amortissement",
			"avances_et_acomptes_net",
			"avances_et_acomptes_net_n1",
		},
		"K": {
			"01",
			"actif",
			"avances_et_acomptes_brut",
			"avances_et_acomptes_amortissement",
			"avances_et_acomptes_net",
			"avances_et_acomptes_net_n1",
		},
	},
	"CS": {
		"C": {
			"01",
			"actif",
			"participations_evaluees_mise_en_equivalence_brut",
			"participations_evaluees_mise_en_equivalence_amortissement",
			"participations_evaluees_mise_en_equivalence_net",
			"participations_evaluees_mise_en_equivalence_net_n1",
		},
		"K": {
			"01",
			"actif",
			"participations_evaluees_mise_en_equivalence_brut",
			"participations_evaluees_mise_en_equivalence_amortissement",
			"participations_evaluees_mise_en_equivalence_net",
			"participations_evaluees_mise_en_equivalence_net_n1",
		},
	},
	"CU": {
		"C": {
			"01",
			"actif",
			"autres_participations_brut",
			"autres_participations_amortissement",
			"autres_participations_net",
			"autres_participations_net_n1",
		},
		"K": {
			"01",
			"actif",
			"autres_participations_brut",
			"autres_participations_amortissement",
			"autres_participations_net",
			"autres_participations_net_n1",
		},
	},
	"BB": {
		"C": {
			"01",
			"actif",
			"creances_rattachees_a_des_participations_brut",
			"creances_rattachees_a_des_participations_amortissement",
			"creances_rattachees_a_des_participations_net",
			"creances_rattachees_a_des_participations_net_n1",
		},
		"K": {
			"01",
			"actif",
			"creances_rattachees_a_des_participations_brut",
			"creances_rattachees_a_des_participations_amortissement",
			"creances_rattachees_a_des_participations_net",
			"creances_rattachees_a_des_participations_net_n1",
		},
	},
	"BD": {
		"C": {
			"01",
			"actif",
			"autres_titres_immobilises_brut",
			"autres_titres_immobilises_amortissement",
			"autres_titres_immobilises_net",
			"autres_titres_immobilises_net_n1",
		},
		"K": {
			"01",
			"actif",
			"autres_titres_immobilises_brut",
			"autres_titres_immobilises_amortissement",
			"autres_titres_immobilises_net",
			"autres_titres_immobilises_net_n1",
		},
	},
	"BF": {
		"C": {
			"01",
			"actif",
			"prets_brut",
			"prets_amortissement",
			"prets_net",
			"prets_net_n1",
		},
		"K": {
			"01",
			"actif",
			"prets_brut",
			"prets_amortissement",
			"prets_net",
			"prets_net_n1",
		},
	},
	"BH": {
		"C": {
			"01",
			"actif",
			"autres_immobilisations_financieres_brut",
			"autres_immobilisations_financieres_amortissement",
			"autres_immobilisations_financieres_net",
			"autres_immobilisations_financieres_net_n1",
		},
		"K": {
			"01",
			"actif",
			"autres_immobilisations_financieres_brut",
			"autres_immobilisations_financieres_amortissement",
			"autres_immobilisations_financieres_net",
			"autres_immobilisations_financieres_net_n1",
		},
	},
	"BJ": {
		"C": {
			"01",
			"actif",
			"total_I_brut",
			"total_I_amortissement",
			"total_I_net",
			"total_I_net_n1",
		},
		"K": {
			"01",
			"actif",
			"total_I_brut",
			"total_I_amortissement",
			"total_I_net",
			"total_I_net_n1",
		},
	},
	"BL": {
		"C": {
			"01",
			"actif",
			"matieres_premieres_approvisionnements_brut",
			"matieres_premieres_approvisionnements_amortissement",
			"matieres_premieres_approvisionnements_net",
			"matieres_premieres_approvisionnements_net_n1",
		},
		"K": {
			"01",
			"actif",
			"matieres_premieres_approvisionnements_brut",
			"matieres_premieres_approvisionnements_amortissement",
			"matieres_premieres_approvisionnements_net",
			"matieres_premieres_approvisionnements_net_n1",
		},
	},
	"BN": {
		"C": {
			"01",
			"actif",
			"en_cours_de_production_de_biens_brut",
			"en_cours_de_production_de_biens_amortissement",
			"en_cours_de_production_de_biens_net",
			"en_cours_de_production_de_biens_net_n1",
		},
		"K": {
			"01",
			"actif",
			"en_cours_de_production_de_biens_brut",
			"en_cours_de_production_de_biens_amortissement",
			"en_cours_de_production_de_biens_net",
			"en_cours_de_production_de_biens_net_n1",
		},
	},
	"BP": {
		"C": {
			"01",
			"actif",
			"en_cours_de_production_de_services_brut",
			"en_cours_de_production_de_services_amortissement",
			"en_cours_de_production_de_services_net",
			"en_cours_de_production_de_services_net_n1",
		},
		"K": {
			"01",
			"actif",
			"en_cours_de_production_de_services_brut",
			"en_cours_de_production_de_services_amortissement",
			"en_cours_de_production_de_services_net",
			"en_cours_de_production_de_services_net_n1",
		},
	},
	"BR": {
		"C": {
			"01",
			"actif",
			"produits_intermediaires_et_finis_brut",
			"produits_intermediaires_et_finis_amortissement",
			"produits_intermediaires_et_finis_net",
			"produits_intermediaires_et_finis_net_n1",
		},
		"K": {
			"01",
			"actif",
			"produits_intermediaires_et_finis_brut",
			"produits_intermediaires_et_finis_amortissement",
			"produits_intermediaires_et_finis_net",
			"produits_intermediaires_et_finis_net_n1",
		},
	},
	"BT": {
		"C": {
			"01",
			"actif",
			"marchandises_brut",
			"marchandises_amortissement",
			"marchandises_net",
			"marchandises_net_n1",
		},
		"K": {
			"01",
			"actif",
			"marchandises_brut",
			"marchandises_amortissement",
			"marchandises_net",
			"marchandises_net_n1",
		},
	},
	"BV": {
		"C": {
			"01",
			"actif",
			"avances_et_acomptes_verses_sur_commandes_brut",
			"avances_et_acomptes_verses_sur_commandes_amortissement",
			"avances_et_acomptes_verses_sur_commandes_net",
			"avances_et_acomptes_verses_sur_commandes_net_n1",
		},
		"K": {
			"01",
			"actif",
			"avances_et_acomptes_verses_sur_commandes_brut",
			"avances_et_acomptes_verses_sur_commandes_amortissement",
			"avances_et_acomptes_verses_sur_commandes_net",
			"avances_et_acomptes_verses_sur_commandes_net_n1",
		},
	},
	"BX": {
		"C": {
			"01",
			"actif",
			"clients_et_comptes_rattaches_brut",
			"clients_et_comptes_rattaches_amortissement",
			"clients_et_comptes_rattaches_net",
			"clients_et_comptes_rattaches_net_n1",
		},
		"K": {
			"01",
			"actif",
			"clients_et_comptes_rattaches_brut",
			"clients_et_comptes_rattaches_amortissement",
			"clients_et_comptes_rattaches_net",
			"clients_et_comptes_rattaches_net_n1",
		},
	},
	"BZ": {
		"C": {
			"01",
			"actif",
			"autres_creances_brut",
			"autres_creances_amortissement",
			"autres_creances_net",
			"autres_creances_net_n1",
		},
		"K": {
			"01",
			"actif",
			"autres_creances_brut",
			"autres_creances_amortissement",
			"autres_creances_net",
			"autres_creances_net_n1",
		},
	},
	"CB": {
		"C": {
			"01",
			"actif",
			"capital_souscrit_et_appele_non_verse_brut",
			"capital_souscrit_et_appele_non_verse_amortissement",
			"capital_souscrit_et_appele_non_verse_net",
			"capital_souscrit_et_appele_non_verse_net_n1",
		},
		"K": {
			"01",
			"actif",
			"capital_souscrit_et_appele_non_verse_brut",
			"capital_souscrit_et_appele_non_verse_amortissement",
			"capital_souscrit_et_appele_non_verse_net",
			"capital_souscrit_et_appele_non_verse_net_n1",
		},
	},
	"CD": {
		"C": {
			"01",
			"actif",
			"valeurs_mobilieres_de_placement_brut",
			"valeurs_mobilieres_de_placement_amortissement",
			"valeurs_mobilieres_de_placement_net",
			"valeurs_mobilieres_de_placement_net_n1",
		},
		"K": {
			"01",
			"actif",
			"valeurs_mobilieres_de_placement_brut",
			"valeurs_mobilieres_de_placement_amortissement",
			"valeurs_mobilieres_de_placement_net",
			"valeurs_mobilieres_de_placement_net_n1",
		},
	},
	"CF": {
		"C": {
			"01",
			"actif",
			"disponibilites_brut",
			"disponibilites_amortissement",
			"disponibilites_net",
			"disponibilites_net_n1",
		},
		"K": {
			"01",
			"actif",
			"disponibilites_brut",
			"disponibilites_amortissement",
			"disponibilites_net",
			"disponibilites_net_n1",
		},
	},
	"CH": {
		"C": {
			"01",
			"actif",
			"charges_constatees_d_avances_brut",
			"charges_constatees_d_avances_amortissement",
			"charges_constatees_d_avances_net",
			"charges_constatees_d_avances_net_n1",
		},
		"K": {
			"01",
			"actif",
			"charges_constatees_d_avances_brut",
			"charges_constatees_d_avances_amortissement",
			"charges_constatees_d_avances_net",
			"charges_constatees_d_avances_net_n1",
		},
	},
	"CJ": {
		"C": {
			"01",
			"actif",
			"total_II_brut",
			"total_II_amortissement",
			"total_II_net",
			"total_II_net_n1",
		},
		"K": {
			"01",
			"actif",
			"total_II_brut",
			"total_II_amortissement",
			"total_II_net",
			"total_II_net_n1",
		},
	},
	"CW": {
		"C": {
			"01",
			"actif",
			"charges_a_repartir_ou_frais_d_emission_d_emprunt_brut",
			"",
			"charges_a_repartir_ou_frais_d_emission_d_emprunt_net",
			"charges_a_repartir_ou_frais_d_emission_d_emprunt_net_n1",
		},
		"K": {
			"01",
			"actif",
			"charges_a_repartir_ou_frais_d_emission_d_emprunt_brut",
			"",
			"charges_a_repartir_ou_frais_d_emission_d_emprunt_net",
			"charges_a_repartir_ou_frais_d_emission_d_emprunt_net_n1",
		},
	},
	"CM": {
		"C": {
			"01",
			"actif",
			"primes_de_remboursements_des_obligations_brut",
			"",
			"primes_de_remboursements_des_obligations_net",
			"primes_de_remboursements_des_obligations_net_n1",
		},
		"K": {
			"01",
			"actif",
			"primes_de_remboursements_des_obligations_brut",
			"",
			"primes_de_remboursements_des_obligations_net",
			"primes_de_remboursements_des_obligations_net_n1",
		},
	},
	"CN": {
		"C": {
			"01",
			"actif",
			"ecarts_de_conversion_V_brut",
			"",
			"ecarts_de_conversion_V_net",
			"ecarts_de_conversion_V_net_n1",
		},
		"K": {
			"01",
			"actif",
			"ecarts_de_conversion_V_brut",
			"ecarts_de_conversion_V_amortissement",
			"",
			"ecarts_de_conversion_V_net_n1",
		},
	},
	"CO": {
		"C": {
			"01",
			"actif",
			"total_general_0_a_V_brut",
			"total_general_0_a_V_amortissement",
			"total_general_0_a_V_net",
			"total_general_0_a_V_net_n1",
		},
		"K": {
			"01",
			"actif",
			"total_general_0_a_V_brut",
			"total_general_0_a_V_amortissement",
			"total_general_0_a_V_net",
			"total_general_0_a_V_net_n1",
		},
	},
	"CP": {
		"C": {
			"01",
			"actif",
			"parts_a_moins_d_un_an_brut",
			"",
			"",
			"",
		},
		"K": {
			"01",
			"actif",
			"parts_a_moins_d_un_an_brut",
			"",
			"",
			"",
		},
	},
	"CR": {
		"C": {
			"01",
			"actif",
			"parts_a_plus_d_un_an_brut",
			"",
			"",
			"",
		},
		"K": {
			"01",
			"actif",
			"parts_a_plus_d_un_an_brut",
			"",
			"",
			"",
		},
	},
	"DA": {
		"C": {
			"02",
			"passif",
			"capital_social_ou_individuel",
			"capital_social_ou_individuel_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"capital_social_ou_individuel",
			"capital_social_ou_individuel_n1",
			"",
			"",
		},
	},
	"DB": {
		"C": {
			"02",
			"passif",
			"primes_d_emission_de_fusion_d_apport",
			"primes_d_emission_de_fusion_d_apport_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"primes_d_emission_de_fusion_d_apport",
			"primes_d_emission_de_fusion_d_apport_n1",
			"",
			"",
		},
	},
	"EK": {
		"C": {
			"02",
			"passif",
			"ecarts_d_equivalence_de_primes_d_emission_de_fusion_d_apport",
			"ecarts_d_equivalence_de_primes_d_emission_de_fusion_d_apport_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"ecarts_d_equivalence_de_primes_d_emission_de_fusion_d_apport",
			"ecarts_d_equivalence_de_primes_d_emission_de_fusion_d_apport_n1",
			"",
			"",
		},
	},
	"DC": {
		"C": {
			"02",
			"passif",
			"ecarts_de_reevaluation",
			"ecarts_de_reevaluation_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"ecarts_de_reevaluation",
			"ecarts_de_reevaluation_n1",
			"",
			"",
		},
	},
	"DD": {
		"C": {
			"02",
			"passif",
			"reserve_legale_1",
			"reserve_legale_1_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"reserve_legale_1",
			"reserve_legale_1_n1",
			"",
			"",
		},
	},
	"DE": {
		"C": {
			"02",
			"passif",
			"reserves_statutaires_ou_contractuelles",
			"reserves_statutaires_ou_contractuelles_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"reserves_statutaires_ou_contractuelles",
			"reserves_statutaires_ou_contractuelles_n1",
			"",
			"",
		},
	},
	"B1": {
		"C": {
			"02",
			"passif",
			"reserve_speciale_des_provisions_pour_fluctuation_des_cours_reserves_statutaires_ou_contractuelles",
			"reserve_speciale_des_provisions_pour_fluctuation_des_cours_reserves_statutaires_ou_contractuelles_n1",
			"",
			"",
		},
	},
	"DF": {
		"C": {
			"02",
			"passif",
			"reserves_reglementees",
			"reserves_reglementees_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"reserves_reglementees",
			"reserves_reglementees_n1",
			"",
			"",
		},
	},
	"EJ": {
		"C": {
			"02",
			"passif",
			"reserve_relative_a_l_achat_d_oeuvres_originales_d_artistes_reservce",
			"",
			"",
			"",
		},
	},
	"DG": {
		"C": {
			"02",
			"passif",
			"autres_reserves",
			"autres_reserves_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"autres_reserves",
			"autres_reserves_n1",
			"",
			"",
		},
	},
	"DH": {
		"C": {
			"02",
			"passif",
			"report_a_nouveau",
			"report_a_nouveau_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"report_a_nouveau",
			"report_a_nouveau_n1",
			"",
			"",
		},
	},
	"DI": {
		"C": {
			"02",
			"passif",
			"resultat_de_l_exercice_benefice_ou_perte",
			"resultat_de_l_exercice_benefice_ou_perte_n1",
			"",
			"",
		},
	},
	"P1": {
		"K": {
			"02",
			"passif",
			"ecarts_de_conversion",
			"ecarts_de_conversion_n1",
			"",
			"",
		},
		"B": {
			"01",
			"passif",
			"dettes_vers_les_etablissements_de_credit",
			"dettes_vers_les_etablissements_de_credit_n1",
			"",
			"",
		},
		"A": {
			"01",
			"passif",
			"capitaux_propres",
			"capitaux_propres_n1",
			"",
			"",
		},
	},
	"P2": {
		"K": {
			"02",
			"passif",
			"resultat_consolide_part_du_groupe",
			"resultat_consolide_part_du_groupe_n1",
			"",
			"",
		},
		"B": {
			"01",
			"passif",
			"comptes_crediteurs_a_la_clientele",
			"comptes_crediteurs_a_la_clientele_n1",
			"",
			"",
		},
		"A": {
			"01",
			"passif",
			"provisions_techniques_brutes",
			"provisions_techniques_brutes_n1",
			"",
			"",
		},
	},
	"P3": {
		"K": {
			"02",
			"passif",
			"autres",
			"autres_n1",
			"",
			"",
		},
		"B": {
			"01",
			"passif",
			"capital_souscrit",
			"capital_souscrit_n1",
			"",
			"",
		},
		"A": {
			"01",
			"passif",
			"total",
			"total_n1",
			"",
			"",
		},
	},
	"P4": {
		"K": {
			"02",
			"passif",
			"ecarts_de_conversion",
			"ecarts_de_conversion_n1",
			"",
			"",
		},
		"B": {
			"01",
			"passif",
			"primes_d_emission",
			"primes_d_emission_n1",
			"",
			"",
		},
	},
	"P5": {
		"K": {
			"02",
			"passif",
			"ecarts_de_conversion_dans_les_reserves",
			"ecarts_de_conversion_dans_les_reserves_n1",
			"",
			"",
		},
		"B": {
			"01",
			"passif",
			"reserves",
			"reserves_n1",
			"",
			"",
		},
	},
	"P6": {
		"K": {
			"02",
			"passif",
			"ecarts_de_conversion_dans_les_resultats",
			"ecarts_de_conversion_dans_les_resultats_n1",
			"",
			"",
		},
		"B": {
			"01",
			"passif",
			"ecarts_de_reevaluation",
			"ecarts_de_reevaluation_n1",
			"",
			"",
		},
	},
	"P7": {
		"K": {
			"02",
			"passif",
			"total_III",
			"total_III_n1",
			"",
			"",
		},
		"B": {
			"01",
			"passif",
			"report_a_nouveau",
			"report_a_nouveau_n1",
			"",
			"",
		},
	},
	"P8": {
		"K": {
			"02",
			"passif",
			"impots_differes",
			"impots_differes_n1",
			"",
			"",
		},
		"B": {
			"01",
			"passif",
			"resultat_de_l_exercice",
			"resultat_de_l_exercice_n1",
			"",
			"",
		},
	},
	"P9": {
		"K": {
			"02",
			"passif",
			"total",
			"total_n1",
			"",
			"",
		},
	},
	"DJ": {
		"C": {
			"02",
			"passif",
			"subventions_investissement",
			"subventions_investissement_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"subventions_investissement",
			"subventions_investissement_n1",
			"",
			"",
		},
	},
	"DK": {
		"C": {
			"02",
			"passif",
			"provisions_reglementees",
			"provisions_reglementees_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"provisions_reglementees",
			"provisions_reglementees_n1",
			"",
			"",
		},
	},
	"DL": {
		"C": {
			"02",
			"passif",
			"total_I",
			"total_I_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"total_I",
			"total_I_n1",
			"",
			"",
		},
	},
	"DM": {
		"C": {
			"02",
			"passif",
			"produit_des_emissions_de_titres_participatifs",
			"produit_des_emissions_de_titres_participatifs_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"produit_des_emissions_de_titres_participatifs",
			"produit_des_emissions_de_titres_participatifs_n1",
			"",
			"",
		},
	},
	"DN": {
		"C": {
			"02",
			"passif",
			"avances_conditionnees",
			"avances_conditionnees_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"avances_conditionnees",
			"avances_conditionnees_n1",
			"",
			"",
		},
	},
	"DO": {
		"C": {
			"02",
			"passif",
			"total_II",
			"total_II_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"total_II",
			"total_II_n1",
			"",
			"",
		},
	},
	"DP": {
		"C": {
			"02",
			"passif",
			"provisions_pour_risques",
			"provisions_pour_risques_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"provisions_pour_risques",
			"provisions_pour_risques_n1",
			"",
			"",
		},
	},
	"DQ": {
		"C": {
			"02",
			"passif",
			"provisions_pour_charges",
			"provisions_pour_charges_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"provisions_pour_charges",
			"provisions_pour_charges_n1",
			"",
			"",
		},
	},
	"DR": {
		"C": {
			"02",
			"passif",
			"total_III",
			"total_III_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"total_IV",
			"total_IV_n1",
			"",
			"",
		},
	},
	"DS": {
		"C": {
			"02",
			"passif",
			"emprunts_obligataires_convertibles",
			"emprunts_obligataires_convertibles_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"emprunts_obligataires_convertibles",
			"emprunts_obligataires_convertibles_n1",
			"",
			"",
		},
	},
	"DT": {
		"C": {
			"02",
			"passif",
			"autres_emprunts_obligataires",
			"autres_emprunts_obligataires_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"autres_emprunts_obligataires",
			"autres_emprunts_obligataires_n1",
			"",
			"",
		},
	},
	"DU": {
		"C": {
			"02",
			"passif",
			"emprunts_et_dettes_aupres_des_etablissements_de_credit_3",
			"emprunts_et_dettes_aupres_des_etablissements_de_credit_3_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"emprunts_et_dettes_aupres_des_etablissements_de_credit_3",
			"emprunts_et_dettes_aupres_des_etablissements_de_credit_3_n1",
			"",
			"",
		},
	},
	"DV": {
		"C": {
			"02",
			"passif",
			"emprunts_et_dettes_financiers_divers_4",
			"emprunts_et_dettes_financiers_divers_4_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"emprunts_et_dettes_financiers_divers_4",
			"emprunts_et_dettes_financiers_divers_4_n1",
			"",
			"",
		},
	},
	"DW": {
		"C": {
			"02",
			"passif",
			"avances_et_acomptes_recus_sur_commandes_en_cours",
			"avances_et_acomptes_recus_sur_commandes_en_cours_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"avances_et_acomptes_recus_sur_commandes_en_cours",
			"avances_et_acomptes_recus_sur_commandes_en_cours_n1",
			"",
			"",
		},
	},
	"DX": {
		"C": {
			"02",
			"passif",
			"dettes_fournisseurs_et_comptes_rattaches",
			"dettes_fournisseurs_et_comptes_rattaches_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"dettes_fournisseurs_et_comptes_rattaches",
			"dettes_fournisseurs_et_comptes_rattaches_n1",
			"",
			"",
		},
	},
	"DY": {
		"C": {
			"02",
			"passif",
			"dettes_fiscales_et_sociales",
			"dettes_fiscales_et_sociales_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"dettes_fiscales_et_sociales",
			"dettes_fiscales_et_sociales_n1",
			"",
			"",
		},
	},
	"DZ": {
		"C": {
			"02",
			"passif",
			"dettes_sur_immobilisations_et_comptes_rattaches",
			"dettes_sur_immobilisations_et_comptes_rattaches_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"dettes_sur_immobilisations_et_comptes_rattaches",
			"dettes_sur_immobilisations_et_comptes_rattaches_n1",
			"",
			"",
		},
	},
	"EA": {
		"C": {
			"02",
			"passif",
			"autres_dettes",
			"autres_dettes_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"autres_dettes",
			"autres_dettes_n1",
			"",
			"",
		},
	},
	"EB": {
		"C": {
			"02",
			"passif",
			"produits_constates_d_avance_2",
			"produits_constates_d_avance_2_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"produits_constates_d_avance_2",
			"produits_constates_d_avance_2_n1",
			"",
			"",
		},
	},
	"EC": {
		"C": {
			"02",
			"passif",
			"total_IV",
			"total_IV_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"total_IV",
			"total_IV_n1",
			"",
			"",
		},
	},
	"ED": {
		"C": {
			"02",
			"passif",
			"V",
			"V_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"V",
			"V_n1",
			"",
			"",
		},
	},
	"EE": {
		"C": {
			"02",
			"passif",
			"total_general_I_a_V",
			"total_general_I_a_V_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"total_general_I_a_V",
			"total_general_I_a_V_n1",
			"",
			"",
		},
	},
	"EF": {
		"C": {
			"02",
			"passif",
			"reserve_reglementee_des_plus_values_de_total_general_I_a_V",
			"reserve_reglementee_des_plus_values_de_total_general_I_a_V_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"reserve_reglementee_des_plus_values_de_total_general_I_a_V",
			"reserve_reglementee_des_plus_values_de_total_general_I_a_V_n1",
			"",
			"",
		},
	},
	"EG": {
		"C": {
			"02",
			"passif",
			"dettes_et_produits_constates_d_avance_a_moins_d_un_an",
			"dettes_et_produits_constates_d_avance_a_moins_d_un_an_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"dettes_et_produits_constates_d_avance_a_moins_d_un_an",
			"dettes_et_produits_constates_d_avance_a_moins_d_un_an_n1",
			"",
			"",
		},
	},
	"EH": {
		"C": {
			"02",
			"passif",
			"concours_bancaires_courants_et_soldes_crediteurs_de_banques_et_ccp_de_dettes_et_produits_constates_d_avances_a_moins_d_un_an",
			"concours_bancaires_courants_et_soldes_crediteurs_de_banques_et_ccp_de_dettes_et_produits_constates_d_avances_a_moins_d_un_an_n1",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"concours_bancaires_courants_et_soldes_crediteurs_de_banques_et_ccp_de_dettes_et_produits_constates_d_avances_a_moins_d_un_an",
			"concours_bancaires_courants_et_soldes_crediteurs_de_banques_et_ccp_de_dettes_et_produits_constates_d_avances_a_moins_d_un_an_n1",
			"",
			"",
		},
	},
	"EI": {
		"C": {
			"02",
			"passif",
			"emprunts_participaltifs_de_dettes_et_produits_constates_d_avances_a_moins_d_un_an_n1",
			"",
			"",
			"",
		},
		"K": {
			"02",
			"passif",
			"emprunts_participaltifs_de_dettes_et_produits_constates_d_avances_a_moins_d_un_an_n1",
			"",
			"",
			"",
		},
	},
	"FA": {
		"C": {
			"03",
			"compte_de_resultat",
			"ventes_de_marchandises_france",
			"ventes_de_marchandises_export",
			"ventes_de_marchandises_total",
			"ventes_de_marchandises_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"ventes_de_marchandises_france",
			"ventes_de_marchandises_export",
			"ventes_de_marchandises_total",
			"ventes_de_marchandises_total_n1",
		},
	},
	"FD": {
		"C": {
			"03",
			"compte_de_resultat",
			"production_vendue_biens_france",
			"production_vendue_biens_export",
			"production_vendue_biens_total",
			"production_vendue_biens_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"production_vendue_biens_france",
			"production_vendue_biens_export",
			"production_vendue_biens_total",
			"production_vendue_biens_total_n1",
		},
	},
	"FG": {
		"C": {
			"03",
			"compte_de_resultat",
			"production_vendue_services_france",
			"production_vendue_services_export",
			"production_vendue_services_total",
			"production_vendue_services_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"production_vendue_services_france",
			"production_vendue_services_export",
			"production_vendue_services_total",
			"production_vendue_services_total_n1",
		},
	},
	"FJ": {
		"C": {
			"03",
			"compte_de_resultat",
			"chiffres_d_affaires_nets_france",
			"chiffres_d_affaires_nets_export",
			"chiffres_d_affaires_nets_total",
			"chiffres_d_affaires_nets_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"chiffres_d_affaires_nets_france",
			"chiffres_d_affaires_nets_export",
			"chiffres_d_affaires_nets_total",
			"chiffres_d_affaires_nets_total_n1",
		},
	},
	"FM": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"production_stockee",
			"production_stockee_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"production_stockee",
			"production_stockee_n1",
		},
	},
	"FN": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"_n1",
			"production_immobilisee_total",
			"production_immobilisee_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"_n1",
			"production_immobilisee_total",
			"production_immobilisee_total_n1",
		},
	},
	"FO": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"_n1",
			"subventions_d_exploitation_total",
			"subventions_d_exploitation_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"_n1",
			"subventions_d_exploitation_total",
			"subventions_d_exploitation_total_n1",
		},
	},
	"FP": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"reprises_sur_amortissements_et_provisions_total",
			"reprises_sur_amortissements_et_provisions_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"reprises_sur_amortissements_et_provisions_total",
			"reprises_sur_amortissements_et_provisions_total_n1",
		},
	},
	"FQ": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"autres_produits_total",
			"autres_produits_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"autres_produits_total",
			"autres_produits_total_n1",
		},
	},
	"FR": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"total_produits_d_exploitation_I_total",
			"total_produits_d_exploitation_I_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"total_produits_d_exploitation_I_total",
			"total_produits_d_exploitation_I_total_n1",
		},
	},
	"FS": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"achats_de_marchandises_y_compris_droits_de_douane_total",
			"achats_de_marchandises_y_compris_droits_de_douane_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"achats_de_marchandises_y_compris_droits_de_douane_total",
			"achats_de_marchandises_y_compris_droits_de_douane_total_n1",
		},
	},
	"FT": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"variation_de_stock_marchandises_total",
			"variation_de_stock_marchandises_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"variation_de_stock_marchandises_total",
			"variation_de_stock_marchandises_total_n1",
		},
	},
	"FU": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"achats_matieres_premiers_et_autres_approvisionnements_total",
			"achats_matieres_premiers_et_autres_approvisionnements_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"achats_matieres_premiers_et_autres_approvisionnements_total",
			"achats_matieres_premiers_et_autres_approvisionnements_total_n1",
		},
	},
	"FV": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"variation_de_stock_matieres_premieres_et_approvisionnements_total",
			"variation_de_stock_matieres_premieres_et_approvisionnements_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"variation_de_stock_matieres_premieres_et_approvisionnements_total",
			"variation_de_stock_matieres_premieres_et_approvisionnements_total_n1",
		},
	},
	"FW": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"autres_achats_et_charges_externes_total",
			"autres_achats_et_charges_externes_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"autres_achats_et_charges_externes_total",
			"autres_achats_et_charges_externes_total_n1",
		},
	},
	"FX": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"impots_taxes_et_versements_assimiles_total",
			"impots_taxes_et_versements_assimiles_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"impots_taxes_et_versements_assimiles_total",
			"impots_taxes_et_versements_assimiles_total_n1",
		},
	},
	"FY": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"salaires_et_traitements_total",
			"salaires_et_traitements_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"salaires_et_traitements_total",
			"salaires_et_traitements_total_n1",
		},
	},
	"FZ": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"charges_sociales",
			"charges_sociales_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"charges_sociales",
			"charges_sociales_n1",
		},
	},
	"GA": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"dotations_d_exploitation_dotation_aux_amortissements_total",
			"dotations_d_exploitation_dotation_aux_amortissements_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"dotations_d_exploitation_dotation_aux_amortissements_total",
			"dotations_d_exploitation_dotation_aux_amortissements_total_n1",
		},
	},
	"GB": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"dotation_d_exploitation_dotations_aux_provisions_total",
			"dotation_d_exploitation_dotations_aux_provisions_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"dotation_d_exploitation_dotations_aux_provisions_total",
			"dotation_d_exploitation_dotations_aux_provisions_total_n1",
		},
	},
	"GC": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"dotation_d_exploitation_sur_actif_circulant_dotations_aux_provisions_total",
			"dotation_d_exploitation_sur_actif_circulant_dotations_aux_provisions_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"dotation_d_exploitation_sur_actif_circulant_dotations_aux_provisions_total",
			"dotation_d_exploitation_sur_actif_circulant_dotations_aux_provisions_total_n1",
		},
	},
	"GD": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"dotation_d_exploitation_pour_risques_et_charges_dotations_aux_provisions_total",
			"dotation_d_exploitation_pour_risques_et_charges_dotations_aux_provisions_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"dotation_d_exploitation_pour_risques_et_charges_dotations_aux_provisions_total",
			"dotation_d_exploitation_pour_risques_et_charges_dotations_aux_provisions_total_n1",
		},
	},
	"GE": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"autres_charges_total",
			"autres_charges_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"autres_charges_total",
			"autres_charges_total_n1",
		},
	},
	"GF": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"total_des_charges_d_exploitation_II_total",
			"total_des_charges_d_exploitation_II_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"total_des_charges_d_exploitation_II_total",
			"total_des_charges_d_exploitation_II_total_n1",
		},
	},
	"GG": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"resultat_d_exploitation_I_II_total",
			"resultat_d_exploitation_I_II_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"resultat_d_exploitation_I_II_total",
			"resultat_d_exploitation_I_II_total_n1",
		},
	},
	"GH": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"benefice_attribue_ou_perte_transferee_III_total",
			"benefice_attribue_ou_perte_transferee_III_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"benefice_attribue_ou_perte_transferee_III_total",
			"benefice_attribue_ou_perte_transferee_III_total_n1",
		},
	},
	"GI": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"perte_supportee_ou_benefice_transfere_IV_total",
			"perte_supportee_ou_benefice_transfere_IV_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"perte_supportee_ou_benefice_transfere_IV_total",
			"perte_supportee_ou_benefice_transfere_IV_total_n1",
		},
	},
	"GJ": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"produits_financiers_de_participations_total",
			"produits_financiers_de_participations_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"produits_financiers_de_participations_total",
			"produits_financiers_de_participations_total_n1",
		},
	},
	"GK": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"produits_des_autres_valeurs_mobilieres_et_creances_de_l_actif_immobilise_total",
			"produits_des_autres_valeurs_mobilieres_et_creances_de_l_actif_immobilise_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"produits_des_autres_valeurs_mobilieres_et_creances_de_l_actif_immobilise_total",
			"produits_des_autres_valeurs_mobilieres_et_creances_de_l_actif_immobilise_total_n1",
		},
	},
	"GL": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"autres_interets_et_produits_assimiles_total",
			"autres_interets_et_produits_assimiles_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"autres_interets_et_produits_assimiles_total",
			"autres_interets_et_produits_assimiles_total_n1",
		},
	},
	"GM": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"reprises_sur_provisions_et_transferts_de_charges_total",
			"reprises_sur_provisions_et_transferts_de_charges_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"reprises_sur_provisions_et_transferts_de_charges_total",
			"reprises_sur_provisions_et_transferts_de_charges_total_n1",
		},
	},
	"GN": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"differences_positives_de_change_total",
			"differences_positives_de_change_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"differences_positives_de_change_total",
			"differences_positives_de_change_total_n1",
		},
	},
	"GO": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"produits_nets_sur_cessions_de_valeurs_mobilieres_de_placement_total",
			"produits_nets_sur_cessions_de_valeurs_mobilieres_de_placement_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"produits_nets_sur_cessions_de_valeurs_mobilieres_de_placement_total",
			"produits_nets_sur_cessions_de_valeurs_mobilieres_de_placement_total_n1",
		},
	},
	"GP": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"total_des_produits_financiers_V_total",
			"_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"total_des_produits_financiers_V_total",
			"_total_n1",
		},
	},
	"GQ": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"dotations_financieres_sur_amortissements_et_provisions_total",
			"dotations_financieres_sur_amortissements_et_provisions_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"dotations_financieres_sur_amortissements_et_provisions_total",
			"dotations_financieres_sur_amortissements_et_provisions_total_n1",
		},
	},
	"GR": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"interets_et_charges_assimilees_total",
			"interets_et_charges_assimilees_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"interets_et_charges_assimilees_total",
			"interets_et_charges_assimilees_total_n1",
		},
	},
	"GS": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"differences_negatives_de_change_total",
			"differences_negatives_de_change_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"differences_negatives_de_change_total",
			"differences_negatives_de_change_total_n1",
		},
	},
	"GT": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"charges_nettes_sur_cessions_de_valeurs_mobilieres_de_placement_total",
			"charges_nettes_sur_cessions_de_valeurs_mobilieres_de_placement_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"charges_nettes_sur_cessions_de_valeurs_mobilieres_de_placement_total",
			"charges_nettes_sur_cessions_de_valeurs_mobilieres_de_placement_total_n1",
		},
	},
	"GU": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"total_des_charges_financieres_VI_total",
			"total_des_charges_financieres_VI_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"total_des_charges_financieres_VI_total",
			"total_des_charges_financieres_VI_total_n1",
		},
	},
	"GV": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"resultat_financier_V_VI_total",
			"resultat_financier_V_VI_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"resultat_financier_V_VI_total",
			"resultat_financier_V_VI_total_n1",
		},
	},
	"GW": {
		"C": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"resultat_courant_avant_impots_total",
			"resultat_courant_avant_impots_total_n1",
		},
		"K": {
			"03",
			"compte_de_resultat",
			"",
			"",
			"resultat_courant_avant_impots_total",
			"resultat_courant_avant_impots_total_n1",
		},
	},
	"HA": {
		"C": {
			"04",
			"compte_de_resultat",
			"produits_exceptionnels_sur_operations_de_gestion_france",
			"produits_exceptionnels_sur_operations_de_gestion_export",
			"",
			"",
		},
		"K": {
			"04",
			"compte_de_resultat",
			"produits_exceptionnels_sur_operations_de_gestion_france",
			"produits_exceptionnels_sur_operations_de_gestion_export",
			"",
			"",
		},
	},
	"HB": {
		"C": {
			"04",
			"compte_de_resultat",
			"produits_exceptionnels_sur_operations_en_capital_france",
			"produits_exceptionnels_sur_operations_en_capital_export",
			"",
			"",
		},
		"K": {
			"04",
			"compte_de_resultat",
			"produits_exceptionnels_sur_operations_en_capital_france",
			"produits_exceptionnels_sur_operations_en_capital_export",
			"",
			"",
		},
	},
	"HC": {
		"C": {
			"04",
			"compte_de_resultat",
			"reprises_sur_provisions_et_transferts_de_charges_france",
			"reprises_sur_provisions_et_transferts_de_charges_export",
			"",
			"",
		},
		"K": {
			"04",
			"compte_de_resultat",
			"reprises_sur_provisions_et_transferts_de_charges_france",
			"reprises_sur_provisions_et_transferts_de_charges_export",
			"",
			"",
		},
	},
	"HD": {
		"C": {
			"04",
			"compte_de_resultat",
			"total_des_produits_exceptionnels_VII_france",
			"total_des_produits_exceptionnels_VII_export",
			"",
			"",
		},
		"K": {
			"04",
			"compte_de_resultat",
			"total_des_produits_exceptionnels_VII_france",
			"total_des_produits_exceptionnels_VII_export",
			"",
			"",
		},
	},
	"HE": {
		"C": {
			"04",
			"compte_de_resultat",
			"charges_exceptionnelles_sur_operations_de_gestion_france",
			"charges_exceptionnelles_sur_operations_de_gestion_export",
			"",
			"",
		},
		"K": {
			"04",
			"compte_de_resultat",
			"charges_exceptionnelles_sur_operations_de_gestion_france",
			"charges_exceptionnelles_sur_operations_de_gestion_export",
			"",
			"",
		},
	},
	"HF": {
		"C": {
			"04",
			"compte_de_resultat",
			"charges_exceptionnelles_sur_operations_en_capital_france",
			"charges_exceptionnelles_sur_operations_en_capital_export",
			"",
			"",
		},
		"K": {
			"04",
			"compte_de_resultat",
			"charges_exceptionnelles_sur_operations_en_capital_france",
			"charges_exceptionnelles_sur_operations_en_capital_export",
			"",
			"",
		},
	},
	"HG": {
		"C": {
			"04",
			"compte_de_resultat",
			"dotations_exceptionnelles_aux_amortissements_et_provisions_france",
			"dotations_exceptionnelles_aux_amortissements_et_provisions_export",
			"",
			"",
		},
		"K": {
			"04",
			"compte_de_resultat",
			"dotations_exceptionnelles_aux_amortissements_et_provisions_france",
			"dotations_exceptionnelles_aux_amortissements_et_provisions_export",
			"",
			"",
		},
	},
	"HH": {
		"C": {
			"04",
			"compte_de_resultat",
			"total_des_charges_exceptionnelles_VIII_france",
			"total_des_charges_exceptionnelles_VIII_export",
			"",
			"",
		},
		"K": {
			"04",
			"compte_de_resultat",
			"total_des_charges_exceptionnelles_VIII_france",
			"total_des_charges_exceptionnelles_VIII_export",
			"",
			"",
		},
	},
	"HI": {
		"C": {
			"04",
			"compte_de_resultat",
			"resultat_exceptionnel_france",
			"resultat_exceptionnel_export",
			"",
			"",
		},
		"K": {
			"04",
			"compte_de_resultat",
			"resultat_exceptionnel_france",
			"resultat_exceptionnel_export",
			"",
			"",
		},
	},
	"HJ": {
		"C": {
			"04",
			"compte_de_resultat",
			"participation_des_salaries_aux_resultats_de_l_entreprise_france",
			"participation_des_salaries_aux_resultats_de_l_entreprise_export",
			"",
			"",
		},
		"K": {
			"04",
			"compte_de_resultat",
			"participation_des_salaries_aux_resultats_de_l_entreprise_france",
			"participation_des_salaries_aux_resultats_de_l_entreprise_export",
			"",
			"",
		},
	},
	"HK": {
		"C": {
			"04",
			"compte_de_resultat",
			"impots_sur_les_benefices_X_france",
			"impots_sur_les_benefices_X_export",
			"",
			"",
		},
		"K": {
			"04",
			"compte_de_resultat",
			"impots_sur_les_benefices_france",
			"impots_sur_les_benefices_export",
			"",
			"",
		},
	},
	"HL": {
		"C": {
			"04",
			"compte_de_resultat",
			"total_des_produits_france",
			"total_des_produits_export",
			"",
			"",
		},
	},
	"HM": {
		"C": {
			"04",
			"compte_de_resultat",
			"total_des_charges_france",
			"total_des_charges_export",
			"",
			"",
		},
	},
	"HN": {
		"C": {
			"04",
			"compte_de_resultat",
			"benefice_ou_perte_france",
			"benefice_ou_perte_export",
			"",
			"",
		},
	},
	"R1": {
		"K": {
			"04",
			"compte_de_resultat",
			"impots_differes_france",
			"impots_differes_export",
			"",
			"",
		},
		"B": {
			"01",
			"compte_de_resultat",
			"interets_et_produits_assimiles",
			"interets_et_produits_assimiles_n1",
			"",
			"",
		},
		"A": {
			"01",
			"compte_de_resultat",
			"primes_cotisations_acquises",
			"primes_cotisations_acquises_n1",
			"",
			"",
		},
	},
	"R2": {
		"K": {
			"04",
			"compte_de_resultat",
			"dotation_aux_amortissements_des_ecarts_d_acquisition_france",
			"dotation_aux_amortissements_des_ecarts_d_acquisition_export",
			"",
			"",
		},
		"B": {
			"01",
			"compte_de_resultat",
			"interets_et_charges_assimilees",
			"interets_et_charges_assimilees_n1",
			"",
			"",
		},
		"A": {
			"01",
			"compte_de_resultat",
			"charges_des_sinistres",
			"charges_des_sinistres_n1",
			"",
			"",
		},
	},
	"R3": {
		"K": {
			"04",
			"compte_de_resultat",
			"resultat_net_des_societes_mises_en_equivalence_france",
			"resultat_net_des_societes_mises_en_equivalence_export",
			"",
			"",
		},
		"B": {
			"01",
			"compte_de_resultat",
			"resultat_de_l_exercice",
			"resultat_de_l_exercice_n1",
			"",
			"",
		},
		"A": {
			"01",
			"compte_de_resultat",
			"resultat_techniques",
			"resultat_techniques_n1",
			"",
			"",
		},
	},
	"R4": {
		"K": {
			"04",
			"compte_de_resultat",
			"resultat_net_des_societes_mises_en_equivalence_france",
			"resultat_net_des_societes_mises_en_equivalence_export",
			"",
			"",
		},
		"A": {
			"01",
			"compte_de_resultat",
			"resultat_de_l_exercice",
			"resultat_de_l_exercice_n1",
			"",
			"",
		},
	},
	"R5": {
		"K": {
			"04",
			"compte_de_resultat",
			"resultat_net_des_entreprises_integrees_france",
			"resultat_net_des_entreprises_integrees_export",
			"",
			"",
		},
	},
	"R6": {
		"K": {
			"04",
			"compte_de_resultat",
			"resultat_groupe_resultat_net_consolide_france",
			"resultat_groupe_resultat_net_consolide_export",
			"",
			"",
		},
	},
	"R7": {
		"K": {
			"04",
			"compte_de_resultat",
			"part_des_interets_minoritaires_resultat_hors_groupe_france",
			"part_des_interets_minoritaires_resultat_hors_groupe_export",
			"",
			"",
		},
	},
	"R8": {
		"K": {
			"04",
			"compte_de_resultat",
			"resultat_net_part_du_groupe_part_de_la_societe_mere_france",
			"resultat_net_part_du_groupe_part_de_la_societe_mere_export",
			"",
			"",
		},
	},
	"HP": {
		"C": {
			"04",
			"compte_de_resultat",
			"renvois_credit_bail_mobilier_france",
			"renvois_credit_bail_mobilier_export",
			"",
			"",
		},
		"K": {
			"04",
			"compte_de_resultat",
			"renvois_credit_bail_mobilier_france",
			"renvois_credit_bail_mobilier_export",
			"",
			"",
		},
	},
	"HQ": {
		"C": {
			"04",
			"compte_de_resultat",
			"renvois_credit_bail_immobilier_france",
			"renvois_credit_bail_immobilier_export",
			"",
			"",
		},
		"K": {
			"04",
			"compte_de_resultat",
			"renvois_credit_bail_immobilier_france",
			"renvois_credit_bail_immobilier_export",
			"",
			"",
		},
	},
	"A1": {
		"C": {
			"04",
			"compte_de_resultat",
			"renvois_transfert_de_charges_france",
			"renvois_transfert_de_charges_export",
			"",
			"",
		},
		"K": {
			"01",
			"actif",
			"ecarts_d_acquisition_brut",
			"ecarts_d_acquisition_amortissement",
			"ecarts_d_acquisition_net",
			"ecarts_d_acquisition_net_n1",
		},
		"B": {
			"01",
			"actif",
			"creances_sur_les_etablissements_de_credit",
			"creances_sur_les_etablissements_de_credit_n1",
			"",
			"",
		},
		"A": {
			"01",
			"actif",
			"placements",
			"placements_n1",
			"",
			"",
		},
	},
	"A2": {
		"C": {
			"04",
			"compte_de_resultat",
			"renvois_cotisations_personnelles_de_l_exploitant_france",
			"renvois_cotisations_personnelles_de_l_exploitant_export",
			"",
			"",
		},
		"K": {
			"01",
			"actif",
			"sur_valeurs_goodwill_brut",
			"sur_valeurs_goodwill_amortissement",
			"sur_valeurs_goodwill_net",
			"sur_valeurs_goodwill_net_n1",
		},
		"B": {
			"01",
			"actif",
			"creances_sur_la_clientele",
			"creances_sur_la_clientele_n1",
			"",
			"",
		},
		"A": {
			"01",
			"actif",
			"total",
			"total_n1",
			"",
			"",
		},
	},
	"A3": {
		"C": {
			"04",
			"compte_de_resultat",
			"renvois_redevances_pour_concessions_de_brevets_de_licences_produits_france",
			"renvois_redevances_pour_concessions_de_brevets_de_licences_produits_export",
			"",
			"",
		},
		"K": {
			"01",
			"actif",
			"difference_de_premiere_consolidation_brut",
			"difference_de_premiere_consolidation_amortissement",
			"difference_de_premiere_consolidation_net",
			"difference_de_premiere_consolidation_net_n1",
		},
		"B": {
			"01",
			"actif",
			"total",
			"total_n1",
			"",
			"",
		},
	},
	"A4": {
		"C": {
			"04",
			"compte_de_resultat",
			"renvois_redevenaces_pour_concessions_de_brevets_de_licenses_charges_france",
			"renvois_redevenaces_pour_concessions_de_brevets_de_licenses_charges_export",
			"",
			"",
		},
		"K": {
			"01",
			"actif",
			"titres_mis_en_equivalence_brut",
			"titres_mis_en_equivalence_amortissement",
			"titres_mis_en_equivalence_net",
			"titres_mis_en_equivalence_net_n1",
		},
	},
	"CZ": {
		"C": {
			"05",
			"immobilisations_augmentations",
			"acquisitions_frais_d_etablissement_et_de_developpement_ou_de_recherche_valeur_brute_debut_exercice",
			"acquisitions_frais_d_etablissement_et_de_developpement_ou_de_recherche_reevaluation",
			"acquisitions_frais_d_etablissement_et_de_developpement_ou_de_recherche_acquisition",
			"",
		},
	},
	"KD": {
		"C": {
			"05",
			"immobilisations_augmentations",
			"acquisitions_total_dont_autres_postes_immobilisations_incorporelles_valeur_brute_debut_exercice",
			"acquisitions_total_dont_autres_postes_immobilisations_incorporelles_reevaluation",
			"acquisitions_total_dont_autres_postes_immobilisations_incorporelles_acquisition",
			"",
		},
	},
	"LN": {
		"C": {
			"05",
			"immobilisations_augmentations",
			"acquisitions_total_immobilisation_corporelles_valeur_brute_debut_exercice",
			"acquisitions_total_immobilisation_corporelles_reevaluation",
			"acquisitions_total_immobilisation_corporelles_acquisition",
			"",
		},
	},
	"LQ": {
		"C": {
			"05",
			"immobilisations_augmentations",
			"acquisitions_total_immobilisations_financieres_valeur_brute_debut_exercice",
			"acquisitions_total_immobilisations_financieres_reevaluation",
			"acquisitions_total_immobilisations_financieres_acquisition",
			"",
		},
	},
	"0G": {
		"C": {
			"05",
			"immobilisations_augmentations",
			"acquisitions_total_general_valeur_brute_debut_exercice",
			"acquisitions_total_general_reevaluation",
			"acquisitions_total_general_acquisition",
			"",
		},
	},
	"IN": {
		"C": {
			"05",
			"immobilisations_diminutions",
			"frais_d_etablissement_et_de_developpement_ou_de_recherche_poste_a_poste",
			"frais_d_etablissement_et_de_developpement_ou_de_recherche_cessions",
			"frais_d_etablissement_et_de_developpement_ou_de_recherche_valeur_brute_fin_exercice",
			"",
		},
	},
	"IO": {
		"C": {
			"05",
			"immobilisations_diminutions",
			"total_dont_autres_postes_d_immobilisations_incorporelles_poste_a_poste",
			"total_dont_autres_postes_d_immobilisations_incorporelles_cessions",
			"total_dont_autres_postes_d_immobilisations_incorporelles_valeur_brute_fin_exercice",
			"",
		},
	},
	"MY": {
		"C": {
			"05",
			"immobilisations_diminutions",
			"virement_postes_immobilisations_poste_a_poste",
			"",
			"",
			"",
		},
	},
	"NC": {
		"C": {
			"05",
			"immobilisations_diminutions",
			"virement_postes_avances_et_acomptes_poste_a_poste",
			"",
			"",
			"",
		},
	},
	"IY": {
		"C": {
			"05",
			"immobilisations_diminutions",
			"total_immobilisations_corporelles_poste_a_poste",
			"total_immobilisations_corporelles_cessions",
			"total_immobilisations_corporelles_valeur_brute_fin_exercice",
			"",
		},
	},
	"I2": {
		"C": {
			"05",
			"immobilisations_diminutions",
			"prets_et_immobilisations_financieres_poste_a_poste",
			"prets_et_immobilisations_financieres_cessions",
			"prets_et_immobilisations_financieres_valeur_brute_fin_exercice",
			"",
		},
	},
	"I3": {
		"C": {
			"05",
			"immobilisations_diminutions",
			"total_immobilisations_financieres_poste_a_poste",
			"total_immobilisations_financieres_cessions",
			"total_immobilisations_financieres_valeur_brute_fin_exercice",
			"",
		},
	},
	"I4": {
		"C": {
			"05",
			"immobilisations_diminutions",
			"total_general_poste_a_poste",
			"total_general_cessions",
			"total_general_valeur_brute_fin_exercice",
			"",
		},
	},
	"CY": {
		"C": {
			"06",
			"amortissement",
			"frais_d_etablissement_et_de_developpement_ou_de_recherche_debut_exercice",
			"frais_d_etablissement_et_de_developpement_ou_de_recherche_augmentation_dotation_exercice",
			"frais_d_etablissement_et_de_developpement_ou_de_recherche_diminution_reprise",
			"frais_d_etablissement_et_de_developpement_ou_de_recherche_fin_exercice",
		},
	},
	"PE": {
		"C": {
			"06",
			"amortissements",
			"total_dont_autres_immobilisations_incorporelles_debut_exercice",
			"total_dont_autres_immobilisations_incorporelles_augmentation_dotation_exercice",
			"total_dont_autres_immobilisations_incorporelles_diminution_reprise",
			"total_dont_autres_immobilisations_incorporelles_fin_exercice",
		},
	},
	"QU": {
		"C": {
			"06",
			"amortissements",
			"total_immobilisations_corporelles_debut_exercice",
			"total_immobilisations_corporelles_augmentation_dotation_exercice",
			"total_immobilisations_corporelles_diminution_reprise",
			"total_immobilisations_corporelles_fin_exercice",
		},
	},
	"0N": {
		"C": {
			"06",
			"amortissements",
			"total_general_debut_exercice",
			"total_general_augmentation_dotation_exercice",
			"total_general_diminution_reprise",
			"total_general_fin_exercice",
		},
	},
	"Z9": {
		"C": {
			"06",
			"amortissement",
			"charges_a_repartir_ou_frais_d_emission_d_emprunt_debut_exercice",
			"charges_a_repartir_ou_frais_d_emission_d_emprunt_augmentation_dotation_exercice",
			"charges_a_repartir_ou_frais_d_emission_d_emprunt_diminution_reprise",
			"charges_a_repartir_ou_frais_d_emission_d_emprunt_fin_exercice",
		},
	},
	"SP": {
		"C": {
			"06",
			"amortissement",
			"mouvement_sur_charges_rep_primes_de_remboursement_des_obligations_debut_exercice",
			"mouvement_sur_charges_rep_primes_de_remboursement_des_obligations_augmentation_dotation_exercice",
			"mouvement_sur_charges_rep_primes_de_remboursement_des_obligations_diminution_reprise",
			"mouvement_sur_charges_rep_primes_de_remboursement_des_obligations_fin_exercice",
		},
	},
	"3X": {
		"C": {
			"06",
			"amortissement",
			"",
			"",
			"",
			"amortissements_derogatoires_fin_exercice",
		},
	},
	"3Z": {
		"C": {
			"07",
			"provisions",
			"total_privisions_reglementees_debut_exercice",
			"total_privisions_reglementees_augmentation_dotation_exercice",
			"total_privisions_reglementees_diminution_reprise",
			"total_privisions_reglementees_fin_exercice",
		},
	},
	"4A": {
		"C": {
			"07",
			"provisions",
			"",
			"",
			"",
			"pour_litiges_fin_exercice",
		},
	},
	"4E": {
		"C": {
			"07",
			"provisions",
			"",
			"",
			"",
			"pour_garanties_donnees_aux_clients_fin_exercice",
		},
	},
	"4J": {
		"C": {
			"07",
			"provisions",
			"",
			"",
			"",
			"pour_perte_sur_marche_a_terme_fin_exercice",
		},
	},
	"4N": {
		"C": {
			"07",
			"provisions",
			"",
			"",
			"",
			"pour_amendes_et_penalites_fin_exercice",
		},
	},
	"4T": {
		"C": {
			"07",
			"provisions",
			"",
			"",
			"",
			"pour_pertes_de_change_fin_exercice",
		},
	},
	"4X": {
		"C": {
			"07",
			"provisions",
			"",
			"",
			"",
			"pour_pensions_et_obligations_similaires_fin_exercice",
		},
	},
	"5B": {
		"C": {
			"07",
			"provisions",
			"",
			"",
			"",
			"pour_impots_fin_exercice",
		},
	},
	"5F": {
		"C": {
			"07",
			"provisions",
			"",
			"",
			"",
			"pour_renouvellement_des_immobilisations_fin_exercice",
		},
	},
	"EO": {
		"C": {
			"07",
			"provisions",
			"",
			"",
			"",
			"pour_gros_entretien_et_grandes_revisions_ou_grosses_reparations_fin_exercice",
		},
	},
	"5R": {
		"C": {
			"07",
			"provisions",
			"pour_charges_sociales_et_fiscales_sur_conges_a_payer_debut_exercice",
			"pour_charges_sociales_et_fiscales_sur_conges_a_payer_augmentation_dotation_exercice",
			"pour_charges_sociales_et_fiscales_sur_conges_a_payer_diminution_reprise",
			"pour_charges_sociales_et_fiscales_sur_conges_a_payer_fin_exercice",
		},
	},
	"5V": {
		"C": {
			"07",
			"provisions",
			"",
			"",
			"",
			"autres_pour_risques_et_charges_fin_exercice",
		},
	},
	"5Z": {
		"C": {
			"07",
			"provisions",
			"total_pour_risques_et_charges_debut_exercice",
			"total_pour_risques_et_charges_augmentation_dotation_exercice",
			"total_pour_risques_et_charges_diminution_reprise",
			"total_pour_risques_et_charges_fin_exercice",
		},
	},
	"6A": {
		"C": {
			"07",
			"provisions",
			"sur_immobilisations_incorporelles_debut_exercice",
			"sur_immobilisations_incorporelles_augmentation_dotation_exercice",
			"sur_immobilisations_incorporelles_diminution_reprise",
			"sur_immobilisations_incorporelles_fin_exercice",
		},
	},
	"6E": {
		"C": {
			"07",
			"provisions",
			"sur_immobilisations_corporelles_debut_exercice",
			"sur_immobilisations_corporelles_augmentation_dotation_exercice",
			"sur_immobilisations_corporelles_diminution_reprise",
			"sur_immobilisations_corporelles_fin_exercice",
		},
	},
	"02": {
		"C": {
			"07",
			"provisions",
			"sur_immobilisations_titres_mis_en_equivalence_debut_exercice",
			"sur_immobilisations_titres_mis_en_equivalence_augmentation_dotation_exercice",
			"sur_immobilisations_titres_mis_en_equivalence_diminution_reprise",
			"sur_immobilisations_titres_mis_en_equivalence_fin_exercice",
		},
	},
	"9U": {
		"C": {
			"07",
			"provisions",
			"",
			"",
			"",
			"sur_immobilisations_titres_de_participation_fin_exercice",
		},
	},
	"06": {
		"C": {
			"07",
			"provisions",
			"sur_immobilisations_autres_immobilisations_financieres_debut_exercice",
			"sur_immobilisations_autres_immobilisations_financieres_augmentation_dotation_exercice",
			"sur_immobilisations_autres_immobilisations_financieres_diminution_reprise",
			"sur_immobilisations_autres_immobilisations_financieres_fin_exercice",
		},
	},
	"6N": {
		"C": {
			"07",
			"provisions",
			"sur_stocks_et_en_cours_debut_exercice",
			"sur_stocks_et_en_cours_augmentation_dotation_exercice",
			"sur_stocks_et_en_cours_diminution_reprise",
			"sur_stocks_et_en_cours_fin_exercice",
		},
	},
	"6T": {
		"C": {
			"07",
			"provisions",
			"sur_comptes_clients_debut_exercice",
			"sur_comptes_clients_augmentation_dotation_exercice",
			"sur_comptes_clients_diminution_reprise",
			"sur_comptes_clients_fin_exercice",
		},
	},
	"6X": {
		"C": {
			"07",
			"provisions",
			"autres_pour_depreciation_debut_exercice",
			"autres_pour_depreciation_augmentation_dotation_exercice",
			"autres_pour_depreciation_diminution_reprise",
			"autres_pour_depreciation_fin_exercice",
		},
	},
	"7B": {
		"C": {
			"07",
			"provisions",
			"total_pour_depreciation_debut_exercice",
			"total_pour_depreciation_augmentation_dotation_exercice",
			"total_pour_depreciation_diminution_reprise",
			"total_pour_depreciation_fin_exercice",
		},
	},
	"7C": {
		"C": {
			"07",
			"provisions",
			"total_general_debut_exercice",
			"total_general_augmentation_dotation_exercice",
			"total_general_diminution_reprise",
			"total_general_fin_exercice",
		},
	},
	"UE": {
		"C": {
			"07",
			"provisions",
			"",
			"dotations_et_reprise_d_exploitation_de_total_general_augmentation_dotation_exercice",
			"dotations_et_reprise_d_exploitation_de_total_general_diminution_reprise",
			"",
		},
	},
	"UG": {
		"C": {
			"07",
			"provisions",
			"",
			"dotations_et_reprise_financieres_de_total_general_augmentation_dotation_exercice",
			"dotations_et_reprise_financieres_de_total_general_diminution_reprise",
			"",
		},
	},
	"UJ": {
		"C": {
			"07",
			"provisions",
			"",
			"dotations_et_reprise_exceptionnelles_de_total_general_augmentation_dotation_exercice",
			"dotations_et_reprise_exceptionnelles_de_total_general_diminution_reprise",
			"",
		},
	},
	"UL": {
		"C": {
			"08",
			"creances_et_dettes",
			"creances_rattachees_a_des_participations_brut",
			"creances_rattachees_a_des_participations_un_an_au_plus",
			"",
			"",
		},
	},
	"UP": {
		"C": {
			"08",
			"creances_et_dettes",
			"prets_brut",
			"prets_un_an_au_plus",
			"",
			"",
		},
	},
	"UT": {
		"C": {
			"08",
			"creances_et_dettes",
			"autres_immobilisations_financieres_brut",
			"autres_immobilisations_financieres_un_an_au_plus",
			"",
			"",
		},
	},
	"VA": {
		"C": {
			"08",
			"creances_et_dettes",
			"clients_douteux_ou_litigieux_brut",
			"clients_douteux_ou_litigieux_un_an_au_plus",
			"",
			"",
		},
	},
	"UX": {
		"C": {
			"08",
			"creances_et_dettes",
			"autres_creances_clients_brut",
			"autres_creances_clients_un_an_au_plus",
			"",
			"",
		},
	},
	"UO": {
		"C": {
			"08",
			"creances_et_dettes",
			"provision_pour_depreciation_anterieurement_constituee_brut",
			"",
			"",
			"",
		},
	},
	"Z1": {
		"C": {
			"08",
			"creances_et_dettes",
			"_creances_representatives_de_titres_pretes_brut",
			"",
			"",
			"",
		},
	},
	"UY": {
		"C": {
			"08",
			"creances_et_dettes",
			"personnel_et_comptes_rattaches_brut",
			"",
			"",
			"",
		},
	},
	"UZ": {
		"C": {
			"08",
			"creances_et_dettes",
			"securite_sociale_autres_organismes_sociaux_brut",
			"",
			"",
			"",
		},
	},
	"VM": {
		"C": {
			"08",
			"creances_et_dettes",
			"impots_sur_les_benefices_brut",
			"",
			"",
			"",
		},
	},
	"VB": {
		"C": {
			"08",
			"creances_et_dettes",
			"tva_brut",
			"",
			"",
			"",
		},
	},
	"VN": {
		"C": {
			"08",
			"creances_et_dettes",
			"autres_impots_taxes_versements_assimiles_brut",
			"",
			"",
			"",
		},
	},
	"VP": {
		"C": {
			"08",
			"creances_et_dettes",
			"divers_brut",
			"",
			"",
			"",
		},
	},
	"VC": {
		"C": {
			"08",
			"creances_et_dettes",
			"groupe_et_associes_brut",
			"",
			"",
			"",
		},
	},
	"VR": {
		"C": {
			"08",
			"creances_et_dettes",
			"debiteurs_divers_dont_creances_relatives_a_des_operations_de_pension_de_titres_brut",
			"",
			"",
			"",
		},
	},
	"VS": {
		"C": {
			"08",
			"creances_et_dettes",
			"charges_constatees_d_avance_brut",
			"",
			"",
			"",
		},
	},
	"VT": {
		"C": {
			"08",
			"creances_et_dettes",
			"total_etat_des_creances_brut",
			"total_etat_des_creances_un_an_au_plus",
			"total_etat_des_creances_montant_de_un_a_cinq_ans",
			"",
		},
	},
	"7Y": {
		"C": {
			"08",
			"creances_et_dettes",
			"autres_emprunts_obligataires_brut_a_un_an_au_plus_brut",
			"autres_emprunts_obligataires_brut_a_un_an_au_plus_un_an_au_plus",
			"autres_emprunts_obligataires_brut_a_un_an_au_plus_montant_de_un_a_cinq_ans",
			"",
		},
	},
	"7Z": {
		"C": {
			"08",
			"creances_et_dettes",
			"autres_emprunts_obligataires_bruts_a_un_an_au_plus_brut",
			"autres_emprunts_obligataires_bruts_a_un_an_au_plus_un_an_au_plus",
			"autres_emprunts_obligataires_bruts_a_un_an_au_plus_montant_de_un_a_cinq_ans",
			"",
		},
	},
	"VG": {
		"C": {
			"08",
			"creances_et_dettes",
			"emprunts_a_1_an_maximum_a_l_origine_brut",
			"emprunts_a_1_an_maximum_a_l_origine_un_an_au_plus",
			"emprunts_a_1_an_maximum_a_l_origine_montant_de_un_a_cinq_ans",
			"",
		},
	},
	"VH": {
		"C": {
			"08",
			"creances_et_dettes",
			"emprunts_a_plus_d_un_an_a_l_origine_brut",
			"emprunts_a_plus_d_un_an_a_l_origine_un_an_au_plus",
			"emprunts_a_plus_d_un_an_a_l_origine_montant_de_un_a_cinq_ans",
			"",
		},
	},
	"8A": {
		"C": {
			"08",
			"creances_et_dettes",
			"emprunts_et_dettes_financieres_divers_brut",
			"emprunts_et_dettes_financieres_divers_un_an_au_plus",
			"emprunts_et_dettes_financieres_divers_montant_de_un_a_cinq_ans",
			"",
		},
	},
	"8B": {
		"C": {
			"08",
			"creances_et_dettes",
			"fournisseurs_et_comptes_rattaches_brut",
			"fournisseurs_et_comptes_rattaches_un_an_au_plus",
			"fournisseurs_et_comptes_rattaches_montant_de_un_a_cinq_ans",
			"",
		},
	},
	"8C": {
		"C": {
			"08",
			"creances_et_dettes",
			"personnel_et_comptes_rattaches_brut",
			"personnel_et_comptes_rattaches_un_an_au_plus",
			"personnel_et_comptes_rattaches_montant_de_un_a_cinq_ans",
			"",
		},
	},
	"8D": {
		"C": {
			"08",
			"creances_et_dettes",
			"securite_sociale_et_autres_organismes_sociaux_brut",
			"securite_sociale_et_autres_organismes_sociaux_un_an_au_plus",
			"securite_sociale_et_autres_organismes_sociaux_montant_de_un_a_cinq_ans",
			"",
		},
	},
	"8E": {
		"C": {
			"08",
			"creances_et_dettes",
			"impots_sur_les_benefices_brut",
			"impots_sur_les_benefices_un_an_au_plus",
			"impots_sur_les_benefices_montant_de_un_a_cinq_ans",
			"",
		},
	},
	"VW": {
		"C": {
			"08",
			"creances_et_dettes",
			"tva_brut",
			"tva_un_an_au_plus",
			"tva_montant_de_un_a_cinq_ans",
			"",
		},
	},
	"VX": {
		"C": {
			"08",
			"creances_et_dettes",
			"obligations_cautionnees_brut",
			"obligations_cautionnees_un_an_au_plus",
			"obligations_cautionnees_montant_de_un_a_cinq_ans",
			"",
		},
	},
	"VQ": {
		"C": {
			"08",
			"creances_et_dettes",
			"autres_impots_taxes_et_assimiles_brut",
			"autres_impots_taxes_et_assimiles_un_an_au_plus",
			"autres_impots_taxes_et_assimiles_montant_de_un_a_cinq_ans",
			"",
		},
	},
	"8J": {
		"C": {
			"08",
			"creances_et_dettes",
			"dettes_sur_immobilisations_et_comptes_rattaches_brut",
			"dettes_sur_immobilisations_et_comptes_rattaches_un_an_au_plus",
			"dettes_sur_immobilisations_et_comptes_rattaches_montant_de_un_a_cinq_ans",
			"dettes_sur_immobilisations_et_comptes_rattaches_montant_a_plus_de_cinq_ans",
		},
	},
	"VI": {
		"C": {
			"08",
			"creances_et_dettes",
			"groupe_et_associes_brut",
			"groupe_et_associes_un_an_au_plus",
			"groupe_et_associes_montant_de_un_a_cinq_ans",
			"groupe_et_associes_montant_a_plus_de_cinq_ans",
		},
	},
	"8K": {
		"C": {
			"08",
			"creances_et_dettes",
			"autres_dont_dettes_relatives_a_des_operations_de_pension_de_titre_brut",
			"autres_dont_dettes_relatives_a_des_operations_de_pension_de_titre_un_an_au_plus",
			"autres_dont_dettes_relatives_a_des_operations_de_pension_de_titre_montant_de_un_a_cinq_ans",
			"autres_dont_dettes_relatives_a_des_operations_de_pension_de_titre_montant_a_plus_de_cinq_ans",
		},
	},
	"Z2": {
		"C": {
			"08",
			"creances_et_dettes",
			"dette_representative_de_titres_empruntes_brut",
			"dette_representative_de_titres_empruntes_un_an_au_plus",
			"dette_representative_de_titres_empruntes_montant_de_un_a_cinq_ans",
			"",
		},
	},
	"8L": {
		"C": {
			"08",
			"creances_et_dettes",
			"produits_constates_d_avance_brut",
			"produits_constates_d_avance_un_an_au_plus",
			"produits_constates_d_avance_montant_de_un_a_cinq_ans",
			"",
		},
	},
	"VY": {
		"C": {
			"08",
			"creances_et_dettes",
			"total_etat_des_dettes_brut",
			"total_etat_des_dettes_un_an_au_plus",
			"total_etat_des_dettes_montant_de_un_a_cinq_ans",
			"total_etat_des_dettes_montant_a_plus_de_cinq_ans",
		},
	},
	"VJ": {
		"C": {
			"08",
			"creances_et_dettes",
			"emprunts_souscrits_en_cours_d_exercice_brut",
			"",
			"",
			"",
		},
	},
	"VK": {
		"C": {
			"08",
			"creances_et_dettes",
			"emprunts_rembourses_en_cours_d_exercice_brut",
			"",
			"",
			"",
		},
	},
	"ZE": {
		"C": {
			"11",
			"affectation_resultat",
			"dividendes",
			"dividendes_n1",
			"",
			"",
		},
	},
	"YQ": {
		"C": {
			"11",
			"affectation_resultat",
			"engagement_de_credit_bail_mobilier",
			"engagement_de_credit_bail_mobilier_n1",
			"",
			"",
		},
	},
	"YR": {
		"C": {
			"11",
			"affectation_resultat",
			"engagement_de_credit_bail_immobilier",
			"engagement_de_credit_bail_immobilier_n1",
			"",
			"",
		},
	},
	"YS": {
		"C": {
			"11",
			"affectation_resultat",
			"effets_portes_a_l_escompte_et_non_echus",
			"effets_portes_a_l_escompte_et_non_echus_n1",
			"",
			"",
		},
	},
	"YT": {
		"C": {
			"11",
			"affectation_resultat",
			"sous_traitance",
			"sous_traitance_n1",
			"",
			"",
		},
	},
	"XQ": {
		"C": {
			"11",
			"affectation_resultat",
			"location_charges_locatives_et_de_copropriete",
			"location_charges_locatives_et_de_copropriete_n1",
			"",
			"",
		},
	},
	"YU": {
		"C": {
			"11",
			"affectation_resultat",
			"personnel_exterieur_a_l_entreprise",
			"personnel_exterieur_a_l_entreprise_n1",
			"",
			"",
		},
	},
	"SS": {
		"C": {
			"11",
			"affectation_resultat",
			"remuneration_d_intermediaire_et_honoraires_hors_retrocessions",
			"remuneration_d_intermediaire_et_honoraires_hors_retrocessions_n1",
			"",
			"",
		},
	},
	"YV": {
		"C": {
			"11",
			"affectation_resultat",
			"retrocessions_d_honoraires_commission_et_courtages",
			"retrocessions_d_honoraires_commission_et_courtages_n1",
			"",
			"",
		},
	},
	"ST": {
		"C": {
			"11",
			"affectation_resultat",
			"autres_comptes",
			"autres_comptes_n1",
			"",
			"",
		},
	},
	"ZJ": {
		"C": {
			"11",
			"affectation_resultat",
			"total_du_poste_correspondant_a_la_ligne_fw_du_tableau_n_2052",
			"total_du_poste_correspondant_a_la_ligne_fw_du_tableau_n_2052_n1",
			"",
			"",
		},
	},
	"YW": {
		"C": {
			"11",
			"affectation_resultat",
			"taxe_professionnelle",
			"taxe_professionnelle_n1",
			"",
			"",
		},
	},
	"9Z": {
		"C": {
			"11",
			"affectation_resultat",
			"autres_impots_taxes_et_versements_assimiles",
			"autres_impots_taxes_et_versements_assimiles_n1",
			"",
			"",
		},
	},
	"YX": {
		"C": {
			"11",
			"affectation_resultat",
			"total_du_poste_correspondant_a_la_ligne_fx_du_tableau_n_2052",
			"total_du_poste_correspondant_a_la_ligne_fx_du_tableau_n_2052_n1",
			"",
			"",
		},
	},
	"YY": {
		"C": {
			"11",
			"affectation_resultat",
			"montant_de_la_tva_collectee",
			"montant_de_la_tva_collectee_n1",
			"",
			"",
		},
	},
	"YZ": {
		"C": {
			"11",
			"affectation_resultat",
			"total_tva_deductible_sur_biens_et_services",
			"total_tva_deductible_sur_biens_et_services_n1",
			"",
			"",
		},
	},
	"YP": {
		"C": {
			"11",
			"affectation_resultat",
			"effectif_moyen_du_personnel",
			"effectif_moyen_du_personnel_n1",
			"",
			"",
		},
	},
	"ZR": {
		"C": {
			"11",
			"affectation_resultat",
			"filiales_et_participations",
			"",
			"",
			"",
		},
	},
	"110": {
		"S": {
			"01",
			"actif",
			"total_general_brut",
			"total_general_amortissement",
			"total_general_net",
			"total_general_net_n1",
		},
	},
	"010": {
		"S": {
			"01",
			"actif",
			"immobilisations_incorporelles_fond_commercial_brut",
			"immobilisations_incorporelles_fond_commercial_amortissement",
			"immobilisations_incorporelles_fond_commercial_net",
			"immobilisations_incorporelles_fond_commercial_net_n1",
		},
	},
	"014": {
		"S": {
			"01",
			"actif",
			"immobilisations_incorporelles_autres_brut",
			"immobilisations_incorporelles_autres_amortissement",
			"immobilisations_incorporelles_autres_net",
			"immobilisations_incorporelles_autres_net_n1",
		},
	},
	"028": {
		"S": {
			"01",
			"actif",
			"immobilisations_corporelles_brut",
			"immobilisations_corporelles_amortissement",
			"immobilisations_corporelles_net",
			"immobilisations_corporelles_net_n1",
		},
	},
	"040": {
		"S": {
			"01",
			"actif",
			"immobilisations_financieres_brut",
			"immobilisations_financieres_amortissement",
			"immobilisations_financieres_net",
			"immobilisations_financieres_net_n1",
		},
	},
	"044": {
		"S": {
			"01",
			"actif",
			"total_actif_immobilise_brut",
			"total_actif_immobilise_amortissement",
			"total_actif_immobilise_net",
			"total_actif_immobilise_net_n1",
		},
	},
	"050": {
		"S": {
			"01",
			"actif",
			"matieres_premiers_approvisionnements_en_cours_de_production_brut",
			"matieres_premiers_approvisionnements_en_cours_de_production_amortissement",
			"matieres_premiers_approvisionnements_en_cours_de_production_net",
			"matieres_premiers_approvisionnements_en_cours_de_production_net_n1",
		},
	},
	"060": {
		"S": {
			"01",
			"actif",
			"stock_marchandises_brut",
			"stock_marchandises_amortissement",
			"stock_marchandises_net",
			"stock_marchandises_net_n1",
		},
	},
	"064": {
		"S": {
			"01",
			"actif",
			"avances_et_acomptes_verses_sur_commandes_brut",
			"avances_et_acomptes_verses_sur_commandes_amortissement",
			"avances_et_acomptes_verses_sur_commandes_net",
			"avances_et_acomptes_verses_sur_commandes_net_n1",
		},
	},
	"068": {
		"S": {
			"01",
			"actif",
			"clients_et_comptes_rattaches_brut",
			"clients_et_comptes_rattaches_amortissement",
			"clients_et_comptes_rattaches_net",
			"clients_et_comptes_rattaches_net_n1",
		},
	},
	"072": {
		"S": {
			"01",
			"actif",
			"creances_autres_brut",
			"creances_autres_amortissement",
			"creances_autres_net",
			"creances_autres_net_n1",
		},
	},
	"080": {
		"S": {
			"01",
			"actif",
			"valeurs_mobilieres_de_placement_brut",
			"valeurs_mobilieres_de_placement_amortissement",
			"valeurs_mobilieres_de_placement_net",
			"valeurs_mobilieres_de_placement_net_n1",
		},
	},
	"084": {
		"S": {
			"01",
			"actif",
			"disponibilites_brut",
			"disponibilites_amortissement",
			"disponibilites_net",
			"disponibilites_net_n1",
		},
	},
	"088": {
		"S": {
			"01",
			"actif",
			"",
			"",
			"caisse_net",
			"caisse_net_n1",
		},
	},
	"092": {
		"S": {
			"01",
			"actif",
			"charges_constatees_d_avance_brut",
			"charges_constatees_d_avance_amortissement",
			"charges_constatees_d_avance_net",
			"charges_constatees_d_avance_net_n1",
		},
	},
	"096": {
		"S": {
			"01",
			"actif",
			"total_actif_circulant_et_charges_constatees_d_avance_brut",
			"total_actif_circulant_et_charges_constatees_d_avance_amortissement",
			"total_actif_circulant_et_charges_constatees_d_avance_net",
			"total_actif_circulant_et_charges_constatees_d_avance_net_n1",
		},
	},
	"120": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"capital_social_ou_individuel_net",
			"capital_social_ou_individuel_net_n1",
		},
	},
	"124": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"ecarts_de_reevaluation_net",
			"ecarts_de_reevaluation_net_n1",
		},
	},
	"126": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"reserve_legale_net",
			"reserve_legale_net_n1",
		},
	},
	"130": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"reserves_reglementees",
			"reserves_reglementees_n1",
		},
	},
	"132": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"autres_reserves",
			"autres_reserves_n1",
		},
	},
	"134": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"report_a_nouveau_net",
			"report_a_nouveau_net_n1",
		},
	},
	"136": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"resultat_de_l_exercice_net",
			"resultat_de_l_exercice_net_n1",
		},
	},
	"140": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"provisions_reglementees_net",
			"provisions_reglementees_net_n1",
		},
	},
	"142": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"total_des_capitaux_propres_I_net",
			"total_des_capitaux_propres_I_net_n1",
		},
	},
	"154": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"provisions_pour_risques_et_charges_II_net",
			"provisions_pour_risques_et_charges_II_net_n1",
		},
	},
	"156": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"emprunts_et_dettes_assimilees_net",
			"emprunts_et_dettes_assimilees_net_n1",
		},
	},
	"164": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"avances_et_acomptes_recus_sur_commandes_en_cours_net",
			"avances_et_acomptes_recus_sur_commandes_en_cours_net_n1",
		},
	},
	"166": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"fournisseurs_et_comptes_rattaches_net",
			"fournisseurs_et_comptes_rattaches_net_n1",
		},
	},
	"169": {
		"S": {
			"01",
			"passif",
			"",
			"comptes_courant_d_associes_de_l_exercice_de_autres_dettes_dont",
			"",
			"",
		},
	},
	"172": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"autres_dettes_net",
			"autres_dettes_net_n1",
		},
	},
	"174": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"produits_constates_d_avance_net",
			"produits_constates_d_avance_net_n1",
		},
	},
	"176": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"total_des_dettes_net",
			"total_des_dettes_net_n1",
		},
	},
	"180": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"total_general_passif_net",
			"total_general_passif_net_n1",
		},
	},
	"193": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"immobilisations_financieres_a_moins_d_un_an_de_total_general_passif_net",
			"",
		},
	},
	"195": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"dettes_a_plus_d_un_an_de_total_general_passif_net",
			"dettes_a_plus_d_un_an_de_total_general_passif_net_n1",
		},
	},
	"197": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"creances_a_plus_d_un_an_de_total_general_passif_net",
			"",
		},
	},
	"182": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"cout_de_revient_des_immobilisations_acquises_ou_creees_au_cours_de_l_exercice_net",
			"cout_de_revient_des_immobilisations_acquises_ou_creees_au_cours_de_l_exercice_net_n1",
		},
	},
	"199": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"comptes_courant_d_associes_debiteurs_de_cout_de_revient_des_immobilisations_acquises_ou_creees_au_cours_de_l_exercice_net",
			"",
		},
	},
	"184": {
		"S": {
			"01",
			"passif",
			"",
			"",
			"prix_de_vente_hors_tva_des_immobilisations_cedees_au_cours_de_l_exercice_net",
			"prix_de_vente_hors_tva_des_immobilisaitons_cedees_au_cours_de_l_exercice_net_n1",
		},
	},
	"209": {
		"S": {
			"02",
			"compte_de_resultat",
			"ventes_de_marchandises_export",
			"",
			"",
			"",
		},
	},
	"210": {
		"S": {
			"02",
			"compte_de_resultat",
			"ventes_de_marchandises_france",
			"",
			"",
			"",
		},
	},
	"214": {
		"S": {
			"02",
			"compte_de_resultat",
			"production_vendue_de_biens_france",
			"_n1",
			"",
			"",
		},
	},
	"215": {
		"S": {
			"02",
			"compte_de_resultat",
			"production_vendue_de_biens_exports",
			"",
			"",
			"",
		},
	},
	"217": {
		"S": {
			"02",
			"compte_de_resultat",
			"production_vendue_de_services_export",
			"production_vendue_de_services_export_n1",
			"",
			"",
		},
	},
	"218": {
		"S": {
			"02",
			"compte_de_resultat",
			"production_vendue_de_services_france",
			"production_vendue_de_services_france_n1",
			"",
			"",
		},
	},
	"222": {
		"S": {
			"02",
			"compte_de_resultat",
			"production_stockee",
			"production_stockee_n1",
			"",
			"",
		},
	},
	"224": {
		"S": {
			"02",
			"compte_de_resultat",
			"production_immobilisee",
			"production_immobilisee_n1",
			"",
			"",
		},
	},
	"226": {
		"S": {
			"02",
			"compte_de_resultat",
			"subventions_d_exploitation_recues",
			"subventions_d_exploitation_recues_n1",
			"",
			"",
		},
	},
	"230": {
		"S": {
			"02",
			"compte_de_resultat",
			"autres_produits",
			"autres_produits_n1",
			"",
			"",
		},
	},
	"232": {
		"S": {
			"02",
			"compte_de_resultat",
			"total_des_produits_d_exploitation_hors_tva",
			"total_des_produits_d_exploitation_hors_tva_n1",
			"",
			"",
		},
	},
	"234": {
		"S": {
			"02",
			"compte_de_resultat",
			"total_des_produits_d_exploitation_hors_tva",
			"total_des_produits_d_exploitation_hors_tva_n1",
			"",
			"",
		},
	},
	"236": {
		"S": {
			"02",
			"compte_de_resultat",
			"variation_de_stock_marchandises",
			"variation_de_stock_marchandises_n1",
			"",
			"",
		},
	},
	"238": {
		"S": {
			"02",
			"compte_de_resultat",
			"achats_de_matieres_premieres_et_autres_approvisionnements_y_compris_droits_de_douane",
			"achats_de_matieres_premieres_et_autres_approvisionnements_y_compris_droits_de_douane_n1",
			"",
			"",
		},
	},
	"240": {
		"S": {
			"02",
			"compte_de_resultat",
			"variation_de_stock_matieres_premieres_et_approvisionnement",
			"variation_de_stock_matieres_premieres_et_approvisionnement_n1",
			"",
			"",
		},
	},
	"242": {
		"S": {
			"02",
			"compte_de_resultat",
			"autres_charges_externes",
			"autres_charges_externes_n1",
			"",
			"",
		},
	},
	"243": {
		"S": {
			"02",
			"compte_de_resultat",
			"taxe_profesionnelle_de_autres_charges_externes",
			"",
			"",
			"",
		},
	},
	"244": {
		"S": {
			"02",
			"compte_de_resultat",
			"impots_taxes_et_versements_assimiles",
			"impots_taxes_et_versements_assimiles_n1",
			"",
			"",
		},
	},
	"250": {
		"S": {
			"02",
			"compte_de_resultat",
			"remunerations_du_personnel",
			"remunerations_du_personnel_n1",
			"",
			"",
		},
	},
	"252": {
		"S": {
			"02",
			"compte_de_resultat",
			"charges_sociales",
			"charges_sociales_n1",
			"",
			"",
		},
	},
	"254": {
		"S": {
			"02",
			"compte_de_resultat",
			"dotations_aux_amortissements",
			"dotations_aux_amortissements_n1",
			"",
			"",
		},
	},
	"256": {
		"S": {
			"02",
			"compte_de_resultat",
			"dotations_aux_provisions",
			"dotations_aux_provisions_n1",
			"",
			"",
		},
	},
	"259": {
		"S": {
			"02",
			"compte_de_resultat",
			"provisions_fiscales_pour_implantations_commerciales_a_l_etranger_de_dotations_aux_provisions",
			"",
			"",
			"",
		},
	},
	"262": {
		"S": {
			"02",
			"compte_de_resultat",
			"autres_charges",
			"autres_charges_n1",
			"",
			"",
		},
	},
	"264": {
		"S": {
			"02",
			"compte_de_resultat",
			"total_des_charges_d_exploitation",
			"total_des_charges_d_exploitation_n1",
			"",
			"",
		},
	},
	"270": {
		"S": {
			"02",
			"compte_de_resultat",
			"resultat_d_exploitation",
			"resultat_d_exploitation_n1",
			"",
			"",
		},
	},
	"280": {
		"S": {
			"02",
			"compte_de_resultat",
			"produits_financiers",
			"produits_financiers_n1",
			"",
			"",
		},
	},
	"290": {
		"S": {
			"02",
			"compte_de_resultat",
			"produits_exceptionnels",
			"produits_exceptionnels_n1",
			"",
			"",
		},
	},
	"294": {
		"S": {
			"02",
			"compte_de_resultat",
			"charges_financiers",
			"charges_financiers_n1",
			"",
			"",
		},
	},
	"300": {
		"S": {
			"02",
			"compte_de_resultat",
			"charges_exceptionnelles",
			"charges_exceptionnelles_n1",
			"",
			"",
		},
	},
	"306": {
		"S": {
			"02",
			"compte_de_resultat",
			"impots_sur_les_benefices",
			"impots_sur_les_benefices_n1",
			"",
			"",
		},
	},
	"310": {
		"S": {
			"02",
			"compte_de_resultat",
			"benefice_ou_perte",
			"benefice_ou_perte_n1",
			"",
			"",
		},
	},
	"316": {
		"S": {
			"02",
			"compte_de_resultat",
			"remuneration_et_avantages_personnels_non_deductibles",
			"remuneration_et_avantages_personnels_non_deductibles_n1",
			"",
			"",
		},
	},
	"374": {
		"S": {
			"02",
			"compte_de_resultat",
			"montant_tva_collectee",
			"montant_tva_collectee_n1",
			"",
			"",
		},
	},
	"376": {
		"S": {
			"02",
			"compte_de_resultat",
			"effectif_moyen_du_personnel",
			"",
			"",
			"",
		},
	},
	"378": {
		"S": {
			"02",
			"compte_de_resultat",
			"montant_de_la_tva_deductible_sur_biens_et_services",
			"montant_de_la_tva_deductible_sur_biens_et_services_n1",
			"",
			"",
		},
	},
	"24B": {
		"S": {
			"02",
			"compte_de_resultat",
			"credit_bail_mobilier_de_montant_de_la_tva_deductible_sur_biens_et_services",
			"",
			"",
			"",
		},
	},
	"24A": {
		"S": {
			"02",
			"compte_de_resultat",
			"credit_bail_immobilier_de_montant_de_la_tva_deductible_sur_biens_et_services",
			"",
			"",
			"",
		},
	},
	"402": {
		"S": {
			"03",
			"immobilisations",
			"augmentations_immobilisations_incorporelles_fond_commercial",
			"",
			"",
			"",
		},
	},
	"404": {
		"S": {
			"03",
			"immobilisations",
			"diminutions_immobilisations_incorporelles_fond_commercial",
			"",
			"",
			"",
		},
	},
	"412": {
		"S": {
			"03",
			"immobilisations",
			"augmentations_immobilisations_incorporelles_autres",
			"",
			"",
			"",
		},
	},
	"414": {
		"S": {
			"03",
			"immobilisations",
			"diminutions_immobilisations_incorporelles_autres_immobilisations_incorporelles",
			"",
			"",
			"",
		},
	},
	"422": {
		"S": {
			"03",
			"immobilisations",
			"augmentations_immobilisations_corporelles_terrains",
			"",
			"",
			"",
		},
	},
	"432": {
		"S": {
			"03",
			"immobilisations",
			"augmentations_immobilisations_corporelles_constructions",
			"",
			"",
			"",
		},
	},
	"442": {
		"S": {
			"03",
			"immobilisations",
			"augmentations_immobilisations_corporelles_installations_techniques_materiel_et_outillage",
			"",
			"",
			"",
		},
	},
	"452": {
		"S": {
			"03",
			"immobilisations",
			"augmentations_immobilisations_corporelles_installations_generales_agencements_divers",
			"",
			"",
			"",
		},
	},
	"462": {
		"S": {
			"03",
			"immobilisations",
			"augmentations_immobilisations_corporelles_materiel_de_transport",
			"",
			"",
			"",
		},
	},
	"472": {
		"S": {
			"03",
			"immobilisations",
			"augmentations_immobilisations_corporelles_autres",
			"",
			"",
			"",
		},
	},
	"482": {
		"S": {
			"03",
			"immobilisations",
			"augmentations_immobilisations_financieres",
			"",
			"",
			"",
		},
	},
	"484": {
		"S": {
			"03",
			"immobilisations",
			"diminutions_immobilisations_financieres",
			"",
			"",
			"",
		},
	},
	"490": {
		"S": {
			"03",
			"immobilisations",
			"total_valeur_brute",
			"",
			"",
			"",
		},
	},
	"492": {
		"S": {
			"03",
			"immobilisations",
			"total_augmentations",
			"",
			"",
			"",
		},
	},
	"494": {
		"S": {
			"03",
			"immobilisations",
			"total_diminutions",
			"",
			"",
			"",
		},
	},
	"582": {
		"S": {
			"03",
			"immobilisations",
			"total_plusvalues_moinsvalues_valeur_residuelle",
			"",
			"",
			"",
		},
	},
	"584": {
		"S": {
			"03",
			"immobilisations",
			"total_plusvalues_moinsvalues_prix_de_cession",
			"",
			"",
			"",
		},
	},
	"585": {
		"S": {
			"03",
			"immobilisations",
			"total_amortissement_plusvalues_moinsvalues_long_terme_19_pourcent",
			"",
			"",
			"",
		},
	},
	"596": {
		"S": {
			"03",
			"immobilisations",
			"total_amortissement_plusvalues_moinsvalues_court_terme",
			"",
			"",
			"",
		},
	},
	"597": {
		"S": {
			"03",
			"immobilisations",
			"total_amortissement_plusvalues_moinsvalues_long_terme_15_ou_12_8_pourcent",
			"",
			"",
			"",
		},
	},
	"599": {
		"S": {
			"03",
			"immobilisations",
			"total_amortissement_plusvalues_moinsvalues_long_terme_0_pourcent",
			"",
			"",
			"",
		},
	},
	// "374": {
	// 	"S": {
	// 		"releve_des_provisions_amortissements_derogatoires",
	// 		"04",
	// 		"divers_montant_de_la_tva_collectee",
	// 		"",
	// 		"",
	// 		"",
	// 	},
	// },
	// "378": {
	// 	"S": {
	// 		"releve_des_provisions_amortissements_derogatoires",
	// 		"04",
	// 		"divers_montant_de_la_tva_deductible_sur_biens_et_services_sauf_immobilisations",
	// 		"",
	// 		"",
	// 		"",
	// 	},
	// },
	"602": {
		"S": {
			"04",
			"releve_des_provisions_amortissements_derogatoires",
			"augmentations_provisions_reglementees_amortissements_derogatoires",
			"",
			"",
			"",
		},
	},
	"603": {
		"S": {
			"04",
			"releve_des_provisions_amortissements_derogatoires",
			"majorations_exceptionnelles_de_augmentations_provisions_reglementees",
			"",
			"",
			"",
		},
	},
	"604": {
		"S": {
			"04",
			"releve_des_provisions_amortissements_derogatoires",
			"diminutions_provisions_reglementees_amortissements_derogatoires",
			"",
			"",
			"",
		},
	},
	"605": {
		"S": {
			"04",
			"releve_des_provisions_amortissements_derogatoires",
			"diminutions_provisions_reglementees_dont_majorations_exceptionnelles_de_30pourcents",
			"",
			"",
			"",
		},
	},
	"612": {
		"S": {
			"04",
			"releve_des_provisions_amortissements_derogatoires",
			"augmentations_provisions_reglementees_autres_provisions_reglementees",
			"",
			"",
			"",
		},
	},
	"614": {
		"S": {
			"04",
			"releve_des_provisions_amortissements_derogatoires",
			"diminutions_provisions_reglementees_autres_provisions_reglementees",
			"",
			"",
			"",
		},
	},
	"622": {
		"S": {
			"04",
			"releve_des_provisions_amortissements_derogatoires",
			"augmentations_provisions_pour_risques_et_charges",
			"",
			"",
			"",
		},
	},
	"624": {
		"S": {
			"04",
			"releve_des_provisions_amortissements_derogatoires",
			"diminutions_provisions_pour_risques_et_charges",
			"",
			"",
			"",
		},
	},
	"632": {
		"S": {
			"04",
			"releve_des_provisions_amortissements_derogatoires",
			"augmentations_provisions_pour_depreciation_sur_immobilisations",
			"",
			"",
			"",
		},
	},
	"634": {
		"S": {
			"04",
			"releve_des_provisions_amortissements_derogatoires",
			"diminutions_provisions_pour_depreciation_sur_immobilisations",
			"",
			"",
			"",
		},
	},
	"642": {
		"S": {
			"04",
			"releve_des_provisions_amortissements_derogatoires",
			"augmentations_provisions_pour_depreciation_sur_stocks_et_en_cours",
			"",
			"",
			"",
		},
	},
	"644": {
		"S": {
			"04",
			"releve_des_provisions_amortissements_derogatoires",
			"diminutions_provisions_pour_depreciation_sur_stock_et_en_cours",
			"",
			"",
			"",
		},
	},
	"652": {
		"S": {
			"04",
			"releve_des_provisions_amortissements_derogatoires",
			"augmentations_provisions_pour_depreciation_sur_clients_et_comptes_rattaches",
			"",
			"",
			"",
		},
	},
	"654": {
		"S": {
			"04",
			"releve_des_provisions_amortissements_derogatoires",
			"diminutions_provisions_pour_depreciation_sur_clients_et_comptes_rattaches",
			"",
			"",
			"",
		},
	},
	"662": {
		"S": {
			"04",
			"releve_des_provisions_amortissements_derogatoires",
			"augmentations_provisions_pour_depreciation_autres_provisions_pour_depreciation",
			"",
			"",
			"",
		},
	},
	"664": {
		"S": {
			"04",
			"releve_des_provisions_amortissements_derogatoires",
			"diminutions_provisions_pour_depreciation_autres_provisions_pour_depreciation",
			"",
			"",
			"",
		},
	},
	"682": {
		"S": {
			"04",
			"releve_des_provisions_amortissements_derogatoires",
			"augmentations_total_releve_des_provisions",
			"",
			"",
			"",
		},
	},
	"684": {
		"S": {
			"04",
			"releve_des_provisions_amortissements_derogatoires",
			"diminutions_total_releve_des_provisions",
			"",
			"",
			"",
		},
	},
}
