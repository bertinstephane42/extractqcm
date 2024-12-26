package main

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// Structure pour les items de question
type Item struct {
	Texte   string `json:"texte"`
	Reponse bool   `json:"reponse"`
}

// Structure pour les questions
type Question struct {
	Question string `json:"question"`
	Option   string `json:"option"`
	Items    []Item `json:"items"`
}

// Structure pour les données contenues dans "donnees"
type Donnees struct {
	Questions []Question `json:"questions"`
}

func showHelp() {
	fmt.Println(`Usage: main [fichier.zip]
Ce programme permet d'extraire un fichier donnees.json contenu dans un fichier .zip généré par le site digistorm.app et d'exporter les données des QCM au format .csv.

Arguments :
  fichier.zip    Chemin vers le fichier ZIP contenant les données du QCM.

Sortie :
  Le programme génère un fichier CSV dans le même dossier que le fichier ZIP, avec le même nom mais une extension .csv.

Exemple :
  main mon_fichier.zip

Encodage :
  - Sous Windows : CRLF est utilisé pour la compatibilité.
  - Sous Linux/Mac : LF est utilisé.`)
}

func main() {
	if len(os.Args) < 2 || os.Args[1] == "--help" || os.Args[1] == "-h" {
		showHelp()
		return
	}

	zipFilename := os.Args[1]
	csvFilename := changeFileExtension(zipFilename, ".csv")

	// Ouvrir le fichier zip
	zipReader, err := zip.OpenReader(zipFilename)
	if err != nil {
		fmt.Printf("Erreur lors de l'ouverture du fichier zip : %v\n", err)
		return
	}
	defer zipReader.Close()

	var jsonData []byte
	for _, file := range zipReader.File {
		if file.Name == "donnees.json" {
			jsonFile, err := file.Open()
			if err != nil {
				fmt.Printf("Erreur lors de l'ouverture de donnees.json : %v\n", err)
				return
			}
			defer jsonFile.Close()

			jsonData, err = io.ReadAll(jsonFile)
			if err != nil {
				fmt.Printf("Erreur lors de la lecture de donnees.json : %v\n", err)
				return
			}
			break
		}
	}

	if jsonData == nil {
		fmt.Println("Fichier donnees.json introuvable dans le zip.")
		return
	}

	// Parser les données JSON
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		fmt.Printf("Erreur lors du parsing du JSON : %v\n", err)
		return
	}

	// Extraire le titre principal
	titre := ""
	if t, ok := data["titre"].(string); ok {
		titre = t
	}

	// Extraire les données des questions
	donneesStr := data["donnees"].(string)
	var donnees Donnees
	if err := json.Unmarshal([]byte(donneesStr), &donnees); err != nil {
		fmt.Printf("Erreur lors du parsing des données des questions : %v\n", err)
		return
	}

	// Créer le fichier CSV
	csvFile, err := os.Create(csvFilename)
	if err != nil {
		fmt.Printf("Erreur lors de la création du fichier CSV : %v\n", err)
		return
	}
	defer csvFile.Close()

	// Ajouter un BOM si le système d'exploitation est Windows
	if runtime.GOOS == "windows" {
		csvFile.Write([]byte{0xEF, 0xBB, 0xBF}) // UTF-8 BOM
	}

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	// Écrire l'en-tête
	header := []string{"Question", "Option", "Texte Réponse", "Réponse Correcte"}
	if err := csvWriter.Write(header); err != nil {
		fmt.Printf("Erreur lors de l'écriture de l'en-tête dans le CSV : %v\n", err)
		return
	}

	// Écrire le titre principal du QCM
	if titre != "" {
		titreRow := []string{titre, "", "", ""}
		if err := csvWriter.Write(titreRow); err != nil {
			fmt.Printf("Erreur lors de l'écriture du titre principal dans le CSV : %v\n", err)
			return
		}
	}

	// Parcourir les questions et écrire dans le CSV
	for _, question := range donnees.Questions {
		for _, item := range question.Items {
			row := []string{question.Question, question.Option, item.Texte, fmt.Sprintf("%v", item.Reponse)}
			if err := csvWriter.Write(row); err != nil {
				fmt.Printf("Erreur lors de l'écriture dans le CSV : %v\n", err)
				return
			}
		}
	}

	fmt.Printf("Exportation terminée : %s\n", csvFilename)
}

// changeFileExtension remplace l'extension d'un fichier par une autre
func changeFileExtension(filename, newExt string) string {
	return filepath.Base(filename[:len(filename)-len(filepath.Ext(filename))]) + newExt
}
