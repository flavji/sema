package reportTemplates


type Section struct {
	Title      string       `firestore:"title"`
	Subsections []string `firestore:"subsections"`
}

type ReportTemplate struct {
	Sections []Section `firestore:"sections"`
}

