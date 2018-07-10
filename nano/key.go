package nano

type Key struct {
	Private string `json:"private"`
	Public  string `json:"public"`
	Account string `json:"account"`
}

func (n *Node) DeterministicKey(seed string, index string) (*Key, error) {
	args := map[string]interface{}{
		"seed":  seed,
		"index": index,
	}
	var nodeResponse Key
	err := n.call("deterministic_key", args, &nodeResponse)
	return &nodeResponse, err
}
