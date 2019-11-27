select 
nom_fichier,
siren,
date_cloture_exercice,
date_cloture_exercice_precedent,
duree_exercice,
code_type_bilan,
compte_de_resultat_chiffres_d_affaires_nets_total,
passif_resultat_de_l_exercice_benefice_ou_perte,
compte_de_resultat_achats_matieres_premiers_et_autres_approvisionnements_total,
compte_de_resultat_impots_taxes_et_versements_assimiles_total,
--  dotation d'exploitation aux amortissements et provisions
coalesce(compte_de_resultat_dotations_d_exploitation_dotation_aux_amortissements_total,0) + coalesce(compte_de_resultat_dotation_d_exploitation_dotations_aux_provisions_total,0) + coalesce(compte_de_resultat_dotation_d_exploitation_sur_actif_circulant_dotations_aux_provisions_total,0) + coalesce(compte_de_resultat_dotation_d_exploitation_pour_risques_et_charges_dotations_aux_provisions_total,0) as dotation_exploitation_amortissements_et_provisions,
-- excedent brut d'exploitation
coalesce(compte_de_resultat_chiffres_d_affaires_nets_total,0) + coalesce(compte_de_resultat_production_stockee, 0) + coalesce(compte_de_resultat_production_immobilisee_total, 0) + coalesce(compte_de_resultat_subventions_d_exploitation_total, 0) - coalesce(compte_de_resultat_achats_de_marchandises_y_compris_droits_de_douane_total,0) - coalesce(compte_de_resultat_variation_de_stock_marchandises_total,0) - coalesce(compte_de_resultat_achats_matieres_premiers_et_autres_approvisionnements_total,0) - coalesce(compte_de_resultat_variation_de_stock_matieres_premieres_et_approvisionnements_total,0) - coalesce(compte_de_resultat_autres_achats_et_charges_externes_total,0) -coalesce(compte_de_resultat_impots_taxes_et_versements_assimiles_total,0) - coalesce(compte_de_resultat_salaires_et_traitements_total,0) - coalesce(compte_de_resultat_charges_sociales,0) as excedent_brut_exploitation,
-- liquidité générale
(1.0 * coalesce(actif_total_II_brut,0) - 1.0 * coalesce(actif_charges_constatees_d_avances_brut, 0) - 1.0 * coalesce(actif_total_II_amortissement, 0) + 1.0 * coalesce(actif_charges_constatees_d_avances_amortissement, 0)) / (1.0 * coalesce(passif_avances_et_acomptes_recus_sur_commandes_en_cours, 0) + 1.0 * coalesce(passif_dettes_fournisseurs_et_comptes_rattaches, 0)+ 1.0 * coalesce(passif_dettes_fiscales_et_sociales, 0)+ 1.0 * coalesce(passif_dettes_sur_immobilisations_et_comptes_rattaches, 0)+ 1.0 * coalesce(passif_autres_dettes, 0) + 1.0 * coalesce(passif_concours_bancaires_courants_et_soldes_crediteurs_de_banques_et_ccp_de_dettes_et_produits_constates_d_avances_a_moins_d_un_an, 0)) as liquidite_generale,
-- liquidité réduite
(1.0 * coalesce(actif_total_II_brut,0) - 1.0 * coalesce(actif_charges_constatees_d_avances_brut, 0) - 1.0 * coalesce(actif_total_II_amortissement, 0) + 1.0 * coalesce(actif_charges_constatees_d_avances_amortissement, 0) - 1.0 * coalesce(actif_matieres_premieres_approvisionnements_brut,0) - 1.0 * coalesce(actif_en_cours_de_production_de_biens_brut,0) - 1.0 * coalesce(actif_en_cours_de_production_de_services_brut,0) - 1.0 * coalesce(actif_produits_intermediaires_et_finis_brut,0) - 1.0 * coalesce(actif_marchandises_brut,0) + 1.0 * coalesce(actif_matieres_premieres_approvisionnements_amortissement,0)
+ 1.0 * coalesce(actif_en_cours_de_production_de_biens_amortissement,0) + 1.0 * coalesce(actif_en_cours_de_production_de_services_amortissement,0) + 1.0 * coalesce(actif_produits_intermediaires_et_finis_amortissement,0) + 1.0 * coalesce(actif_marchandises_amortissement,0)) / (1.0 * coalesce(passif_avances_et_acomptes_recus_sur_commandes_en_cours, 0) + 1.0 * coalesce(passif_dettes_fournisseurs_et_comptes_rattaches, 0)+ 1.0 * coalesce(passif_dettes_fiscales_et_sociales, 0)+ 1.0 * coalesce(passif_dettes_sur_immobilisations_et_comptes_rattaches, 0)+ 1.0 * coalesce(passif_autres_dettes, 0) + 1.0 * coalesce(passif_concours_bancaires_courants_et_soldes_crediteurs_de_banques_et_ccp_de_dettes_et_produits_constates_d_avances_a_moins_d_un_an, 0)) as liquidite_reduite,
-- autonomie financière
(1.0 * coalesce(passif_total_I, 0) + 1.0 * coalesce(passif_total_II, 0) ) * 100 / (1.0 * coalesce(passif_total_general_I_a_V, 0)) as autonomie_financiere,
-- taux d'intérêt financier
(1.0 * coalesce(compte_de_resultat_interets_et_charges_assimilees_total,0) * 100) / (1.0 * coalesce(compte_de_resultat_chiffres_d_affaires_nets_total,0)) as taux_interet_financier,
-- poids besoin en fond de roulement sur le chiffre d'affaire
(1.0 * coalesce(actif_matieres_premieres_approvisionnements_brut, 0) 
+ 1.0 * coalesce(actif_en_cours_de_production_de_biens_brut, 0)
+ 1.0 * coalesce(actif_en_cours_de_production_de_services_brut, 0) 
+ 1.0 * coalesce(actif_produits_intermediaires_et_finis_brut, 0) 
+ 1.0 * coalesce(actif_marchandises_brut, 0) 
+ 1.0 * coalesce(actif_avances_et_acomptes_verses_sur_commandes_brut, 0) 
+ 1.0 * coalesce(actif_clients_et_comptes_rattaches_brut, 0) 
+ 1.0 * coalesce(actif_autres_creances_brut, 0) 
+ 1.0 * coalesce(actif_capital_souscrit_et_appele_non_verse_brut, 0) 
- (
  1.0 * coalesce(actif_matieres_premieres_approvisionnements_amortissement, 0) 
  + 1.0 * coalesce(actif_en_cours_de_production_de_biens_amortissement, 0) 
  + 1.0 * coalesce(actif_en_cours_de_production_de_services_amortissement, 0) 
  + 1.0 * coalesce(actif_produits_intermediaires_et_finis_amortissement, 0) 
  + 1.0 * coalesce(actif_marchandises_amortissement, 0) 
  + 1.0 * coalesce(actif_avances_et_acomptes_verses_sur_commandes_amortissement, 0) 
  + 1.0 * coalesce(actif_clients_et_comptes_rattaches_amortissement, 0) 
  + 1.0 * coalesce(actif_autres_creances_amortissement, 0) 
  + 1.0 * coalesce(actif_capital_souscrit_et_appele_non_verse_amortissement, 0)
) 
+ 1.0 * coalesce(actif_charges_constatees_d_avances_brut, 0) 
- 1.0 * coalesce(actif_charges_constatees_d_avances_amortissement, 0) 
+ 1.0 * coalesce(affectation_resultat_effets_portes_a_l_escompte_et_non_echus, 0)
) * 100 /
(1.0 * coalesce(compte_de_resultat_chiffres_d_affaires_nets_total, 0)) as poids_besoins_en_fond_de_roulement_sur_ca,
-- part des salariés
(1.0 * coalesce(compte_de_resultat_salaires_et_traitements_total, 0) +
1.0 * coalesce(compte_de_resultat_charges_sociales, 0) +
1.0 * coalesce(compte_de_resultat_participation_des_salaries_aux_resultats_de_l_entreprise_france, 0)) * 100/
(
  1.0 * coalesce(compte_de_resultat_chiffres_d_affaires_nets_total, 0)
+ 1.0 * coalesce(compte_de_resultat_production_stockee,0)
+ 1.0 * coalesce(compte_de_resultat_production_immobilisee_total,0)
- 1.0 * coalesce(compte_de_resultat_achats_de_marchandises_y_compris_droits_de_douane_total,0)
- 1.0 * coalesce(compte_de_resultat_variation_de_stock_marchandises_total,0)
- 1.0 * coalesce(compte_de_resultat_achats_matieres_premiers_et_autres_approvisionnements_total,0)
- 1.0 * coalesce(compte_de_resultat_variation_de_stock_matieres_premieres_et_approvisionnements_total,0)
- 1.0 * coalesce(compte_de_resultat_autres_achats_et_charges_externes_total,0)
) as part_salaries,
( 1.0 * coalesce(compte_de_resultat_resultat_courant_avant_impots_total,0)
- 1.0 * coalesce(compte_de_resultat_reprises_sur_amortissements_et_provisions_total,0)
+ 1.0 * coalesce(compte_de_resultat_dotations_d_exploitation_dotation_aux_amortissements_total,0)
+ 1.0 * coalesce(compte_de_resultat_dotation_d_exploitation_dotations_aux_provisions_total,0)
+ 1.0 * coalesce(compte_de_resultat_dotation_d_exploitation_sur_actif_circulant_dotations_aux_provisions_total,0)
+ 1.0 * coalesce(compte_de_resultat_dotation_d_exploitation_pour_risques_et_charges_dotations_aux_provisions_total,0)
- 1.0 * coalesce(compte_de_resultat_reprises_sur_provisions_et_transferts_de_charges_total,0)
+ 1.0 * coalesce(compte_de_resultat_dotations_financieres_sur_amortissements_et_provisions_total,0)
+ 1.0 * coalesce(compte_de_resultat_produits_exceptionnels_sur_operations_de_gestion_export,0)
- 1.0 * coalesce(compte_de_resultat_charges_exceptionnelles_sur_operations_de_gestion_export,0)
- 1.0 *  coalesce(compte_de_resultat_participation_des_salaries_aux_resultats_de_l_entreprise_export,0)
- 1.0 * coalesce(compte_de_resultat_impots_sur_les_benefices_X_export,0)
) * 100/ (1.0 * coalesce(compte_de_resultat_chiffres_d_affaires_nets_total, 0)
+ 1.0 * coalesce(compte_de_resultat_production_stockee,0)
+ 1.0 * coalesce(compte_de_resultat_production_immobilisee_total,0)
- 1.0 * coalesce(compte_de_resultat_achats_de_marchandises_y_compris_droits_de_douane_total,0)
- 1.0 * coalesce(compte_de_resultat_variation_de_stock_marchandises_total,0)
- 1.0 * coalesce(compte_de_resultat_achats_matieres_premiers_et_autres_approvisionnements_total,0)
- 1.0 * coalesce(compte_de_resultat_variation_de_stock_matieres_premieres_et_approvisionnements_total,0)
- 1.0 * coalesce(compte_de_resultat_autres_achats_et_charges_externes_total,0)
)  as part_autofinancement
,(1.0 * coalesce(compte_de_resultat_impots_taxes_et_versements_assimiles_total,0)
 + 1.0 * coalesce(compte_de_resultat_impots_sur_les_benefices_X_france,0)
 ) * 100 /
( 1.0 * coalesce(compte_de_resultat_chiffres_d_affaires_nets_total, 0)
+ 1.0 * coalesce(compte_de_resultat_production_stockee,0)
 + 1.0 * coalesce(compte_de_resultat_production_immobilisee_total,0)
 - 1.0 * coalesce(compte_de_resultat_achats_de_marchandises_y_compris_droits_de_douane_total,0)
 - 1.0 * coalesce(compte_de_resultat_variation_de_stock_marchandises_total,0)
 - 1.0 * coalesce(compte_de_resultat_achats_matieres_premiers_et_autres_approvisionnements_total,0)
 - 1.0 * coalesce(compte_de_resultat_variation_de_stock_matieres_premieres_et_approvisionnements_total,0)
 - 1.0 * coalesce(compte_de_resultat_autres_achats_et_charges_externes_total,0)) as part_etat,
 1.0 * coalesce(compte_de_resultat_resultat_courant_avant_impots_total,0)
* 100  / (
1.0 * coalesce(compte_de_resultat_chiffres_d_affaires_nets_total, 0)
+ 1.0 * coalesce(compte_de_resultat_subventions_d_exploitation_total, 0)
) as performance
from bilan