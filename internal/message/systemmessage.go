package message

// SysCreateNewNode will be sent through the ovneto tool to create a new node.
// The message will first go to the nodemgmt, which will then create the node instance.
type SysCreateNewNode struct {
	Username             string `json:"user"`
	DepthVision          uint8  `json:"depthVision"`
	CurrentNodeIp        string `json:"ip"`
	ConnectionNodeIp     string `json:"connIp"`
	Port                 uint16 `json:"port"`
	ConnectionNodePort   uint16 `json:"connPort"`
	UseDns               bool   `json:"dns"`
	ConnectionsCapacity  uint16 `json:"connCap"`
	MessageQueueCapacity uint16 `json:"msgCap"`
	IsNewNode            bool   `json:"new"`
	NetworkName          string `json:"network"`
	Help                 bool   `json:"help"`
}

// SysStartNode will be sent through the ovneto tool to start a node instance that exists.
// The message will first go to the nodemgmt, which will then look for the node instance.
// If it is not running, nodemgmt will start it.
type SysStartNode struct {
	Username string `json:"user"`
	Network  string `json:"network"`
}

// SysStopNode will be sent through the ovneto tool to stop a node instance that is running.
// The message will first go to the nodemgmt, which will then look for the node instance.
// If it is running, nodemgmt will stop it.
type SysStopNode struct {
	Username string `json:"user"`
	Network  string `json:"network"`
}
