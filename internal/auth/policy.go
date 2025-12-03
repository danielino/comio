package auth

// Policy represents an authorization policy
type Policy struct {
	Version   string
	Statement []Statement
}

// Statement represents a policy statement
type Statement struct {
	Effect    string
	Action    []string
	Resource  []string
	Condition map[string]interface{}
}
