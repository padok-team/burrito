package comment

type Comment interface {
	Generate(string) (string, error)
}
