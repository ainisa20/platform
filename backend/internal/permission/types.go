package permission

// Manifest declares all permissions for one system (platform or shop).
type Manifest struct {
	SystemType string
	Nodes      []Node
}

// Node represents a directory, menu, or the container for button permissions.
type Node struct {
	Name      string
	Type      int16 // PermTypeDirectory / PermTypeMenu / PermTypeButton
	Path      string
	Component string
	Icon      string
	Sort      int
	Children  []Node
	Buttons   []Button
}

// Button represents a single button-level permission under a menu node.
type Button struct {
	Name string
	Code string
}
