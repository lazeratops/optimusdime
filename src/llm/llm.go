package llm

type Llm interface {
	FindElements(elements DesiredElements, content string) (map[string]int, error)
}

type Config struct {
	ApiKey string
	ApiUrl string
}

type Name = string
type Description = string

type DesiredElements map[Name]Description
