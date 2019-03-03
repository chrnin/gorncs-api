# gorncs-api 

Outil de synchronisation et d'APIfication dédié au Registre National du Commerce et des Sociétés

L'exploitation des bilans du RNCS nécessite que vous obteniez une license auprès de l'INPI
Pour plus d'information: https://www.inpi.fr/fr/licence-registre-national-du-commerce-et-des-societes-rncs

Les accès au serveur FTPS ainsi que les conditions de cette accès seront explicités dans les documents de retour.  
## fonctionnalités
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
```
~/go/bin/gorncs-api 
gorncs-api écoute 127.0.0.1:3000
Pour plus d'information: gorncs-api --help
```
### Utilisation
Ce microservice prend un numéro siren en paramètre dans l'URL et retourne l'ensemble des bilans dans un tableau JSON où chaque élément est un objet dont les clés font référence aux notions que l'on retrouve sur les formulaires de déclaration fiscale de la DGFIP. (2030-sd, 2031-sd etc.)

```
$ http :3000/bilan/012345678
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Wed, 27 Feb 2019 14:49:07 GMT
Transfer-Encoding: chunked

[
    {
        "actif_autres_creances_brut": xxxxx,
        "actif_autres_creances_net": xxxxx,
        ...
```

### Modèle de données
WIP

## Problèmes connus
- le modèle de données n'est pas optimal, certains champs sont en doublons et/ou demandent de l'analyse pour faire baisser le nombre de champs
- pas de gestion automatique de la planification du clonage et de l'importation (i.e. pour le moment il faut le faire avec cron)
- documentation du modèle de données à créer, en donnant les références vers les documents déclaratifs DGFIP.

## Feuille de route
- clarifier/documenter le modèle de données
- modèle de fichier excel en lien avec la base de données
- microclient web pour l'API
- microservice d'aggregation et de requête dans la base
- gestion d'authentification
- calculs des principaux ratios d'analyse financière
