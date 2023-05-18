package constant

type Category string

type PathIndex struct {
	FirstIdx, SecondIdx int
}

var Command = map[string]PathIndex{
	"cp":      {1, 2},
	"rm":      {1, -1},
	"upload":  {1, 2},
	"mkdir":   {1, -1},
	"cat":     {1, -1},
	"cd":      {1, -1},
	"chmod":   {1, -1},
	"migrate": {1, -1},
}

var CommandPubsub = map[string]bool{
	"cp":     true,
	"rm":     true,
	"upload": true,
	"mkdir":  true,
	"chmod":  true,
}
