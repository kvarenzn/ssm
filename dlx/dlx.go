// A go implementation of Knuth's Dancing Links algorithm
// reference: sage/combinat/matrices/dancing_links_c.h in project sagemath

package dlx

type Node interface {
	Left() Node
	Right() Node
	Up() Node
	Down() Node
	Header() Node
	Tag() int
	SetLeft(Node)
	SetRight(Node)
	SetUp(Node)
	SetDown(Node)
	SetHeader(Node)
	SetTag(int)
}

type One struct {
	left   Node
	right  Node
	up     Node
	down   Node
	column Node
	tag    int
}

func (o *One) Left() Node {
	return o.left
}

func (o *One) Right() Node {
	return o.right
}

func (o *One) Up() Node {
	return o.up
}

func (o *One) Header() Node {
	return o.column
}

func (o *One) Tag() int {
	return o.tag
}

func (o *One) Down() Node {
	return o.down
}

func (o *One) SetLeft(n Node) {
	o.left = n
}

func (o *One) SetRight(n Node) {
	o.right = n
}

func (o *One) SetUp(n Node) {
	o.up = n
}

func (o *One) SetDown(n Node) {
	o.down = n
}

func (o *One) SetHeader(n Node) {
	o.column = n
}

func (o *One) SetTag(i int) {
	o.tag = i
}

type Column struct {
	One
	Index int
	Ones  int
}

type DLXMode uint8

const (
	DMSearchUnknown DLXMode = iota
	DMSearchForward
	DMSearchAdvance
	DMSearchBackup
	DMSearchRecover
	DMSearchDone
)

type DLX struct {
	Root    Node
	Columns []*Column
	Ones    []*One

	Mode        DLXMode
	CurrentNode Node
	BestColumn  *Column
	Choice      []Node
	Solution    []int
}

func (d *DLX) SmallestColumn() *Column {
	minimal := -1
	var result *Column
	for p := d.Root.Right(); p != d.Root; p = p.Right() {
		if minimal == -1 || p.(*Column).Ones < minimal {
			minimal = p.(*Column).Ones
			result = p.(*Column)
		}
	}

	return result
}

func (*DLX) Remove(column *Column) {
	left := column.Left()
	right := column.Right()
	left.SetRight(right)
	right.SetLeft(left)

	for row := column.Down(); row != column; row = row.Down() {
		for n := row.Right(); n != row; n = n.Right() {
			up := n.Up()
			down := n.Down()

			up.SetDown(down)
			down.SetUp(up)

			n.Header().(*Column).Ones--
		}
	}
}

func (d *DLX) RemoveOthers(n Node) {
	for p := n.Right(); p != n; p = p.Right() {
		d.Remove(p.Header().(*Column))
	}
}

func (*DLX) Recover(column *Column) {
	for row := column.Up(); row != column; row = row.Up() {
		for n := row.Left(); n != row; n = n.Left() {
			up := n.Up()
			down := n.Down()

			up.SetDown(n)
			down.SetUp(n)

			n.Header().(*Column).Ones++
		}
	}

	left := column.left
	right := column.right
	left.SetRight(column)
	right.SetLeft(column)
}

func (d *DLX) RecoverOthers(n Node) {
	for p := n.Left(); p != n; p = p.Left() {
		d.Recover(p.Header().(*Column))
	}
}

func (d *DLX) AddRow(row []int, id int) {
	var rowStart Node

	for _, i := range row {
		n := &One{
			tag:    id,
			column: d.Columns[i+1],
		}
		d.Ones = append(d.Ones, n)

		if rowStart == nil {
			rowStart = n
		}

		n.down = d.Columns[i+1]

		if d.Columns[i+1].Ones == 0 {
			n.up = d.Columns[i+1]
			d.Columns[i+1].down = n
		} else {
			n.up = d.Columns[i+1].up
			d.Columns[i+1].up.SetDown(n)
		}

		d.Columns[i+1].up = n

		if rowStart != n {
			n.left = rowStart.Left()
			n.right = rowStart

			rowStart.Left().SetRight(n)
			rowStart.SetLeft(n)
		} else {
			n.left = n
			n.right = n
		}

		d.Columns[i+1].Ones++
	}
}

func (d *DLX) Search() bool {
	if len(d.Columns) == 0 {
		return false
	}

	if d.Mode == DMSearchDone {
		return false
	}

	if d.CurrentNode != nil || d.BestColumn != nil {
		d.Mode = DMSearchRecover
	} else {
		d.Mode = DMSearchForward
	}

	for {
		if d.Mode == DMSearchForward {
			d.BestColumn = d.SmallestColumn()
			d.Remove(d.BestColumn)

			d.CurrentNode = d.BestColumn.down
			d.Choice = append(d.Choice, d.CurrentNode)

			d.Mode = DMSearchAdvance
		}

		if d.Mode == DMSearchAdvance {
			if d.CurrentNode == d.BestColumn {
				d.Mode = DMSearchBackup
				continue
			}

			d.RemoveOthers(d.CurrentNode)

			if d.Columns[0].right == d.Columns[0] {
				d.Solution = nil
				for _, n := range d.Choice {
					d.Solution = append(d.Solution, n.Tag())
				}

				return true
			}

			d.Mode = DMSearchForward
			continue
		}

		if d.Mode == DMSearchBackup {
			d.Recover(d.BestColumn)
			if len(d.Choice) == 1 {
				d.Mode = DMSearchDone
				continue
			}

			d.Choice = d.Choice[:len(d.Choice)-1]

			d.CurrentNode = d.Choice[len(d.Choice)-1]
			d.BestColumn = d.CurrentNode.Header().(*Column)

			d.Mode = DMSearchRecover
		}

		if d.Mode == DMSearchRecover {
			d.RecoverOthers(d.CurrentNode)

			d.Choice = d.Choice[:len(d.Choice)-1]
			d.CurrentNode = d.CurrentNode.Down()
			d.Choice = append(d.Choice, d.CurrentNode)

			d.Mode = DMSearchAdvance
			continue
		}

		if d.Mode == DMSearchDone {
			return false
		}
	}
}

func NewDLX(rows [][]int) *DLX {
	dlx := &DLX{}

	columns := -1
	for _, row := range rows {
		for _, unit := range row {
			if columns < unit {
				columns = unit
			}
		}
	}
	columns++

	if columns <= 0 {
		return dlx
	}

	dlx.Root = &Column{
		One: One{
			tag: 0,
		},
		Index: 0,
	}
	dlx.Columns = append(dlx.Columns, dlx.Root.(*Column))

	for i := 0; i < columns; i++ {
		col := &Column{
			One: One{
				tag: -1,
			},
			Ones:  0,
			Index: i + 1,
		}

		col.up = col
		col.down = col

		dlx.Root.SetLeft(col)
		dlx.Columns[len(dlx.Columns)-1].SetRight(col)

		col.left = dlx.Columns[len(dlx.Columns)-1]
		col.right = dlx.Root

		dlx.Columns = append(dlx.Columns, col)
	}

	for i, row := range rows {
		dlx.AddRow(row, i)
	}

	return dlx
}
