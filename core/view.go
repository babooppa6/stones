package core

// We use these tables to cheaply approximate FoV, but we cache the tables so
// we only have to compute them once.
var tableCache = make(map[int]map[Offset]map[Offset]struct{})

// FoV uses a simple heuristic to approximate shadowcasting field of view
// calculation. The offsets in the resulting field are reletive to the given
// origin.
func FoV(origin *Tile, radius int) map[Offset]*Tile {
	// Retrieve (or create and cache) the table for the given radius.
	// This table maps a particular offset to a set of offsets which can seen
	// if the given one is transparent. Using this table, we basically just do
	// a recursive search using the table to guide us. Thus, we get a field of
	// view algorithm which performs minimal computation, never revisits tiles,
	// and short circuits on closed maps.
	table, cached := tableCache[radius]
	if !cached {
		table = computeTable(radius)
		tableCache[radius] = table
	}

	fov := map[Offset]*Tile{Offset{0, 0}: origin}
	stack := []Offset{{0, 0}}

	for len(stack) > 0 {
		// Pop an offset from the stack
		off := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		// Get the tile  for that offset
		tile := fov[off]

		for adj := range table[off] {
			// Add all the adjacent tiles to the field of view.
			neighbor := tile.Adjacent[adj.Sub(off)]
			fov[adj] = neighbor

			// If the neighbor is translucient, push it onto the stack to
			// continue exploration. Since we already added it to fov, when we
			// pop it, we'll be able to access the position again.
			if neighbor.Lite {
				stack = append(stack, adj)
			}
		}
	}

	// fix some artifacts related to standing next to a long wall.
	wallfix(fov, radius)
	return fov
}

// computeTable gets the table for a particular radius. This table will allow
// us to approxmiate shadowcasting using FoV.
func computeTable(radius int) map[Offset]map[Offset]struct{} {
	table := make(map[Offset]map[Offset]struct{})

	// We start at the origin, and will compute a single octant.
	addEntry(table, Offset{0, 0}, Offset{1, 0})
	addEntry(table, Offset{0, 0}, Offset{1, 1})

	// The following algorithm is better described in the blog post at:
	// http://stonesrl.blogspot.com/2013/02/pre-computed-fov.html
	// Basically there is a pattern in which tiles spawn both diagonally and
	// horizontally. Each row, the distance between these tiles increases by 1.
	// Everything below such a tile continues diagoanlly, Everything else goes
	// horizontally. A picture is worth a thousand words, so check out the
	// blog post...
	currBreak := 0
	breakCount := 0
	for x := 1; x < radius; x++ {
		nextY := 0
		for y := 0; y <= x; y++ {
			pos := Offset{x, y}
			if y == currBreak {
				addEntry(table, pos, Offset{x + 1, nextY})
				addEntry(table, pos, Offset{x + 1, nextY + 1})
				nextY += 2
			} else {
				addEntry(table, pos, Offset{x + 1, nextY})
				nextY++
			}
		}
		breakCount--
		if breakCount < 0 {
			breakCount = currBreak + 1
			currBreak++
		}
	}

	// Now that we've computed one octant, reflect and rotate to complete the
	// other 7 octants.
	completeTable(table)

	return table
}

// addEntry places a link between two offsets, adding the set keyed by src
// if it is not already present.
func addEntry(table map[Offset]map[Offset]struct{}, src, dst Offset) {
	neighbors, ok := table[src]
	if !ok {
		neighbors = map[Offset]struct{}{}
		table[src] = neighbors
	}
	neighbors[dst] = struct{}{}
}

// completeTable uses reflection and rotation to take a table with a single
// octant and extend it to all 8 octants.
func completeTable(table map[Offset]map[Offset]struct{}) {
	for key := range table {
		from := Offset{key.Y, key.X}
		for pos := range table[key] {
			addEntry(table, from, Offset{pos.Y, pos.X})
		}
	}

	for key := range table {
		negX := Offset{-key.X, key.Y}
		negY := Offset{key.X, -key.Y}
		negXY := Offset{-key.X, -key.Y}
		for pos := range table[key] {
			addEntry(table, negX, Offset{-pos.X, pos.Y})
			addEntry(table, negY, Offset{pos.X, -pos.Y})
			addEntry(table, negXY, Offset{-pos.X, -pos.Y})
		}
	}
}

// wallfix fills in some missing wall artifacts in a field of view.
func wallfix(fov map[Offset]*Tile, radius int) {
	// Each of the four code block does the same basic thing in a different
	// direction. Basically we just march in an orthogonal direction adding
	// adjacent tiles as we go. Each block starts one tile from the origin and
	// goes until either the end of the fov range, which will either be after an
	// non-translucient Tile or the edge of the fov radius. In the case of an
	// non-translucient  Tile, we also check adjacency through the previous
	// (translucient) Tile, avoiding visual inconsistencies when multiple edges
	// connect to a single non-translucient Tile.
	for dx := 1; dx <= radius; dx++ {
		if _, ok := fov[Offset{dx, 0}]; ok {
			pos := fov[Offset{dx - 1, 0}]
			fov[Offset{dx, 1}] = pos.Adjacent[Offset{1, 1}]
			fov[Offset{dx, -1}] = pos.Adjacent[Offset{1, -1}]
		} else {
			break
		}
	}
	for dx := -1; dx >= -radius; dx-- {
		if _, ok := fov[Offset{dx, 0}]; ok {
			pos := fov[Offset{dx + 1, 0}]
			fov[Offset{dx, 1}] = pos.Adjacent[Offset{-1, 1}]
			fov[Offset{dx, -1}] = pos.Adjacent[Offset{-1, -1}]
		} else {
			break
		}
	}
	for dy := 1; dy <= radius; dy++ {
		if _, ok := fov[Offset{0, dy}]; ok {
			pos := fov[Offset{0, dy - 1}]
			fov[Offset{1, dy}] = pos.Adjacent[Offset{1, 1}]
			fov[Offset{-1, dy}] = pos.Adjacent[Offset{-1, 1}]
		} else {
			break
		}
	}
	for dy := -1; dy >= -radius; dy-- {
		if _, ok := fov[Offset{0, dy}]; ok {
			pos := fov[Offset{0, dy + 1}]
			fov[Offset{1, dy}] = pos.Adjacent[Offset{1, -1}]
			fov[Offset{-1, dy}] = pos.Adjacent[Offset{-1, -1}]
		} else {
			break
		}
	}
}

// We use these tables to cheaply approximate LoS, but we cache the tables so
// we only have to compute them once. They are computed by reversing FoV tables
// stored in tableCache.
var reverseTableCache = make(map[int]map[Offset]Offset)

// Trace computes a line of Offset from the origin to the goal Offset.
// The line is computed using the same heuristic as FoV, although it is
// computed without reguard to line of sight.
func Trace(goal Offset) []Offset {
	// setup book keeping
	table := getReverseTable(goal)
	path := []Offset{}

	// compute the path
	for goal.X != 0 || goal.Y != 0 {
		path = append(path, goal)
		goal = table[goal]
	}

	// reverse the path
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	// return path
	return path
}

// LoS returns true if the line from origin to goal computed by Trace does not
// contain a non-translucient Tile. The line is computed using the same
// heuristic as FoV, so if LoS returns true, then the goal tile would also be
// included in the computed field of view (assuming large enough radius).
func LoS(origin, goal *Tile) bool {
	curr := goal.Offset.Sub(origin.Offset)
	table := getReverseTable(curr)
	for goal != origin {
		if !goal.Lite {
			return false
		}
		next := table[curr]
		goal = goal.Adjacent[next.Sub(curr)]
		curr = next
	}
	return true
}

// getReverseTable gets a FoV table and reverses it for LoS computations.
func getReverseTable(o Offset) map[Offset]Offset {
	radius := o.Chebyshev()
	table, cached := reverseTableCache[radius]
	if !cached {
		table = computeReverseTable(radius)
		reverseTableCache[radius] = table
	}
	return table
}

// computeReverseTable computes a LoS table by reversing a FoV table.
func computeReverseTable(radius int) map[Offset]Offset {
	forward, cached := tableCache[radius]
	if !cached {
		forward = computeTable(radius)
		tableCache[radius] = forward
	}

	reverse := make(map[Offset]Offset)
	for pos, edges := range forward {
		for edge := range edges {
			reverse[edge] = pos
		}
	}
	return reverse
}

// TODO Add circular version of FoV
