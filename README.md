# gorncs-api 

API pour le Registre National du Commerce et des Sociétés

L'exploitation des bilans du RNCS nécessite que vous obteniez une license auprès de l'INPI
Pour plus d'information: https://www.inpi.fr/fr/licence-registre-national-du-commerce-et-des-societes-rncs

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
- `go get github.com/signaux-faibles/gorncs-api`

## Utilisation 
Lancé sans argument, gorncs-api ouvre un point d'appel sur 127.0.0.1:3000.  

Afin de peupler la base de données, il faut cloner le dépot [RNCS de l'INPI](https://www.inpi.fr/fr/licence-registre-national-du-commerce-et-des-societes-rncs).

Si le même chemin est configuré pour l'import, les doublons ne seront pas intégrés, il est donc possible de lancer plusieurs fois la même arborescence.

Avant de démarrer l'import, il faut initialiser le schéma de la base avec -initdb.

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

### Exemple
```
$ go get github.com/signaux-faibles/gorncs-api
$ ~/go/bin/gorncs-api -initdb
2019/03/03 11:46:18 initialisation de la base de données Sqlite pour gorncs: ./bilan.db
2019/03/03 11:46:18 creation de la table bilan (858 champs): ok
2019/03/03 11:46:18 creation index: ok
$ mkdir inpi
$ ~/go/bin/gorncs-api -download -user secretUser -password secretPassword -path inpi
2019/03/03 11:49:43 ftp://opendata-rncs.inpi.fr/public/Bilans_Donnees_Saisies/parcours du dossier 
[...]
$ ~/go/bin/gorncs-api -scan -path inpi -limit 1000
2019/03/03 11:52:23 gorncs - analyse de l'arborescence INPI dans inpi
2019/03/03 11:52:23 Bilans importés: 1000
```
## Appel de l'api
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

## Problèmes connus
- le modèle de données n'est pas optimal, certains champs sont en doublons et/ou demandent de l'analyse pour faire baisser le nombre de champs
- pas de gestion automatique de la planification du clonage et de l'importation (i.e. il faudra le faire avec cron)

## Feuille de route
- clarifier le modèle de données
- modèle de fichier excel en lien avec la base de données
- client web pour l'API
- microservice d'aggregation et de requête dans la base
- gestion d'authentification 
