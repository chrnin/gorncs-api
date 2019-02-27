# gorncs

Package pour lire les fichiers au format RNCS de l'INPI

## Usage à partir d'un fichier XML
```go
file, _ := ioutil.ReadFile(filename)
bilan := gorncs.ParseBilan(file)
```

## Usage à partir du dépot RNCS de l'INPI
```go
for bilan := range gorncs.BilanWorker(path) {
 	fmt.Println(bilan)
}
```

## Améliorations nécessaires
- Le schéma des fichiers XML outrepasse la documentation, faire en sorte que les fichiers incorrects soient malgré tout analysés correctement
- améliorer le schéma en sortie, nommage et structure