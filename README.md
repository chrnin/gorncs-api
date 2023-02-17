# gorncs-api 

Outil de synchronisation et d'APIfication dédié au Registre National du Commerce et des Sociétés

L'exploitation des bilans du RNCS nécessite que vous obteniez une licence auprès de l'INPI

Pour plus d'informations: https://data.inpi.fr/content/editorial/Serveur_ftp_entreprises

## Fonctionnalités
- clonage et importation du dépot de fichiers RNCS de l'INPI
- stockage des bilans sous sqlite (exploitable sous Excel par ODBC)
- distribution des bilans et comptes de résultat dans un microservice REST/JSON

## Dépendances

gorncs-api dépend de:

- [Gin Web Framework](http://github.com/gin-gonic/gin)
- [gorncs](http://github.com/chrnin/gorncs)
- [go-sqlite3](http://github.com/mattn/go-sqlite3)
- [go-curl](github.com/andelf/go-curl)
- [libcurl](https://curl.haxx.se/libcurl/)

## Installer gorncs-api
- Installer libcurl4-openssl-dev
- Installer sqlite3
- `go get github.com/signaux-faibles/gorncs-api`

Il faut s'attendre à un temps de compilation assez important du fait de la compilation de go-sqlite3.

## Usage
```
$ gorncs-api -help
Usage of ./gorncs-api:
  -DB string
    	chemin de la base sqlite3 (default "./bilan.db")
  -bind string
    	port d'écoute de l'api (default "127.0.0.1:3000")
  -download
    	synchroniser le dépôt RNCS dans (voir -path, -user et -password)
  -initdb
    	créer une nouvelle base sqlite
  -limit int
    	limiter l'import à n bilans
  -password string
    	mot de passe FTPS RNCS/inpi
  -path string
    	chemin où sont stockés les fichiers RNCS (default ".")
  -scan
    	importer les fichiers
  -siren string
    	restreint l'importation au siren
  -user string
    	utilisateur FTPS RNCS/inpi
  -verbose
    	afficher les informations d'importation
```

### Première synchronisation
Pour une première utilisation, il faut initialiser le schéma dans une nouvelle base de données:
```
$ ~/go/bin/gorncs-api -initdb
2019/03/03 11:46:18 initialisation de la base de données Sqlite pour gorncs: ./bilan.db
2019/03/03 11:46:18 creation de la table bilan (858 champs): ok
2019/03/03 11:46:18 creation index: ok
```
Il faut également prévoir un répertoire vide (le créer au besoin) pour stocker le mirroir du FTPS de l'INPI.

```
$ mkdir /foo/inpi
$ ~/go/bin/gorncs-api -download -user secretUser -password secretPassword -path /foo/inpi
2019/03/03 11:49:43 ftp://opendata-rncs.inpi.fr/public/Bilans_Donnees_Saisies/parcours du dossier 
[...]
```

Une fois la première synchronisation effectuée (dans la limite du quota fixé par la license INPI), il faut scanner l'arborescence pour en extraire les bilans.
```
$ ~/go/bin/gorncs-api -scan -path /foo/inpi
2019/03/03 11:52:23 gorncs - analyse de l'arborescence INPI dans inpi
2019/03/03 11:52:23 Bilans importés: 143512
```

### Synchronisation journalière
Une fois le dépot initialisé et la base de données en place, il est possible de renouveler l'opération pour récupérer les fichiers encore non synchronisés.  
L'INPI a fixé un quota de téléchargement de 1Go/jour, il faut donc s'attendre à plusieurs jours de téléchargement important au départ.  
L'INPI publie les nouveaux bilans à raison d'un fichier de quelques Mo par jour, une planification à 24 heures des deux commandes suivantes permet d'obtenir les mises à jour en suivant leur rythme de publication.
```
$ ~/go/bin/gorncs-api -download -user secretUser -password secretPassword -path /foo/inpi
$ ~/go/bin/gorncs-api -scan -path /foo/inpi
```

### Gestion des erreurs
gorncs-api valide le md5 des fichiers téléchargés, en cas de différence, les fichiers sont supprimés et seront re-téléchargés à la tentative suivante.  

Cette vérification survient également pour tous les fichiers à chaque synchronisation de sorte que si un fichier subit une corruption de ses données il sera supprimé et re-téléchargé durant la synchronisation.  

Chaque tentative de synchronisation ne procède qu'à un seul essais de téléchargement de fichier pour éviter des blocages liés à une corruption venant directement du dépot inpi. Les fichiers ne correspondant pas à leur md5 ne sont pas conservés.

Si une erreur de téléchargement survient lors de la synchronisation, celle-ci est immédiatement arrêtée étant entendu que dans la majorité des cas il s'agit du quota qui est atteint.

Si le quota de téléchargement le permet, il n'y a aucune contre-indication à lancer deux synchronisations successives.

## API
### Lancement
Le plus simplement du monde: 
```
~/go/bin/gorncs-api 
gorncs-api écoute 127.0.0.1:3000
Pour plus d'information: gorncs-api --help
```
### Points d'appel (voir également modèle de données)
#### GET /bilan/:siren
retourne un tableau des bilans connus pour un siren
#### GET /fields/:field
retourne les informations sur un champ. 
#### GET /fields
retourne les informations sur tous les champs.
#### GET /schema
retourne le schéma paramétré dans gorncs-api

### Modèle de données
Le format des bilans est un tableau d'objets comportant les informations des bilans.  
#### Exemple
```JSON
$ http :3000/bilan/015551278
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Mon, 04 Mar 2019 07:44:56 GMT
Transfer-Encoding: chunked

[
    {
        "actif_autres_creances_brut": 170897,
        "actif_autres_creances_net": 170897,
        "actif_autres_creances_net_n1": 91023,
        "actif_autres_immobilisations_corporelles_amortissement": 527712,
        "actif_autres_immobilisations_corporelles_brut": 576353,
        "actif_autres_immobilisations_corporelles_net": 48641,
        ...
    }
]
```

#### Champs permanents
- id integer
- nom_fichier text
- siren text
- date_cloture_exercice datetime
- code_greffe text
- num_depot text
- num_gestion text
- code_activite text
- date_cloture_exercice_precedent datetime
- duree_exercice text
- duree_exercice_precedent text
- date_depot datetime
- code_motif text
- code_type_bilan text
- code_devise text
- code_origine_devise text
- code_confidentialite text
- denomination text
- adresse text
- rapport_integration: permet de savoir si des champs ont été ignorés pendant l'intégration du fichier

#### Information sur les champs comptables
La grande quantité de champs disponibles dans le schéma m'interdit d'en mettre le détail exhaustif dans cette documentation, toutefois, il est possible d'obtenir des informations sur les champs avec le service /fields de l'API:
```JSON
$ http :3000/fields/actif_total_I_net
HTTP/1.1 200 OK
Content-Length: 115
Content-Type: application/json; charset=utf-8
Date: Mon, 04 Mar 2019 07:54:20 GMT

[
    {
        "bilan": "complet",
        "code": "BJ",
        "colonne": 3,
        "page": "01"
    },
    {
        "bilan": "consolide",
        "code": "BJ",
        "colonne": 3,
        "page": "01"
    }
]

```

Il est ainsi possible de connaître pour chaque champ la position d'origine de l'information issue des formulaires de déclaration fiscale (i.e. https://www.impots.gouv.fr/portail/formulaire/2033-sd/bilan-simplifie).  
Un appel à /field sans paramètre retournera le descriptif de tous les champs.

#### Schéma
Il est possible d'obtenir le descriptif du schéma exploité par gorncs-api pour produire la liste des champs avec le service /schema  
Le retour 
```JSON
$ http :3000/schema
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Mon, 04 Mar 2019 08:00:15 GMT
Transfer-Encoding: chunked

{
    "010": {                                                                  | code poste
        "S": [                                                                | code bilan
            "01",                                                             | page
            "actif",                                                          | catégorie
            "immobilisations_incorporelles_fond_commercial_brut",             | rubriques
            "immobilisations_incorporelles_fond_commercial_amortissement",    |
            "immobilisations_incorporelles_fond_commercial_net",              |
            "immobilisations_incorporelles_fond_commercial_net_n1"            |
        ]
    },
    "014": {
        ...
    }
}
```
## Problèmes connus
- le modèle de données n'est pas optimal, certains champs sont en doublons et/ou demandent de l'analyse pour faire baisser le nombre de champs
- pas de gestion automatique de la planification du clonage et de l'importation (i.e. pour le moment il faut le faire avec cron)
- documentation du modèle de données à créer, en donnant les références vers les documents déclaratifs DGFIP.
- pas de tests

## Feuille de route
- clarifier/documenter le modèle de données
- modèle de fichier excel en lien avec la base de données
- microclient web pour l'API
- microservice d'aggregation et de requête dans la base
- gestion d'authentification
- calculs des principaux ratios d'analyse financière
