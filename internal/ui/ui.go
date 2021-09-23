package ui

import (
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/lusingander/fynegit/internal/graph"
	"github.com/lusingander/fynegit/internal/repository"
	"github.com/lusingander/gogigu"
)

const (
	dateTimeFormat   = "2006/01/02 15:04:05"
	messageMaxLength = 80
)

var (
	defaultWindowSize = fyne.NewSize(1400, 800)
)

type manager struct {
	fyne.Window
	rm *repository.RepositoryManager

	*commitDetailView
	*sideMenuView
}

func Start(w fyne.Window, rm *repository.RepositoryManager) {
	m := &manager{
		Window:           w,
		rm:               rm,
		commitDetailView: nil,
	}
	m.SetMainMenu(m.buildMainMenu())
	m.SetContent(m.buildContent())
	m.Resize(defaultWindowSize)
	m.ShowAndRun()
}

func (m *manager) buildContent() fyne.CanvasObject {
	if m.rm == nil {
		return m.buildEmptyView()
	}

	vs := container.NewVSplit(
		m.buildCommitGraphView(),
		m.buildCommitDetailView(),
	)
	vs.SetOffset(0.6)
	hs := container.NewHSplit(
		m.buildSideMenuView(),
		vs,
	)
	hs.SetOffset(0.15)
	return hs
}

func (m *manager) buildMainMenu() *fyne.MainMenu {
	openMenuItem := fyne.NewMenuItem("Open...", m.showRepositoryOpenDialog)
	closeMenuItem := fyne.NewMenuItem("Close repository", m.closeRepository)
	fileMenu := fyne.NewMenu("File", openMenuItem, fyne.NewMenuItemSeparator(), closeMenuItem)
	return fyne.NewMainMenu(fileMenu)
}

func (m *manager) buildEmptyView() fyne.CanvasObject {
	openButton := widget.NewButtonWithIcon(
		"Open Git Repository",
		theme.StorageIcon(),
		m.showRepositoryOpenDialog,
	)
	return container.NewCenter(openButton)
}

func (m *manager) showRepositoryOpenDialog() {
	callback := func(lu fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, m.Window)
			return
		}
		if lu == nil {
			return // canceled
		}
		rm, err := repository.OpenGitRepository(lu.String()[7:]) // `file://`
		if err != nil {
			dialog.ShowError(err, m.Window)
			return
		}
		m.rm = rm
		m.SetContent(m.buildContent())
	}
	dialog.ShowFolderOpen(callback, m.Window)
}

func (m *manager) closeRepository() {
	if m.rm == nil {
		return
	}
	m.rm = nil
	m.SetContent(m.buildContent())
}

func (m *manager) buildCommitGraphView() fyne.CanvasObject {
	if m.rm == nil {
		log.Fatalln("m.rm must not be nil")
	}
	list := widget.NewList(
		func() int {
			return len(m.rm.Nodes)
		},
		func() fyne.CanvasObject {
			return commitGraphItem(m.rm)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			updateCommitGraphItem(m.rm, m.rm.Nodes[id], item)
		},
	)
	list.OnSelected = m.updateCommitDetailView
	return list
}

func commitGraphItem(rm *repository.RepositoryManager) fyne.CanvasObject {
	graphAreaWidth := graph.CalcCommitGraphAreaWidth(rm)
	graphArea := widget.NewLabel("")
	msg := widget.NewLabel("commit message")
	hash := widget.NewLabel("hash")
	author := widget.NewLabel("author")
	committedAt := widget.NewLabel("2006/01/02 15:04:05")
	var msgW, hashW, authorW float32 = 500, 100, 200
	graphArea.Move(fyne.NewPos(0, 0))
	msg.Move(fyne.NewPos(graphArea.Position().X+graphAreaWidth, 0))
	hash.Move(fyne.NewPos(msg.Position().X+msgW, 0))
	author.Move(fyne.NewPos(hash.Position().X+hashW, 0))
	committedAt.Move(fyne.NewPos(author.Position().X+authorW, 0))
	return container.NewWithoutLayout(
		graphArea,
		msg,
		hash,
		author,
		committedAt,
	)
}

func updateCommitGraphItem(rm *repository.RepositoryManager, node *gogigu.Node, item fyne.CanvasObject) {
	objs := item.(*fyne.Container).Objects
	objs[0] = graph.CalcCommitGraphTreeRow(rm, node, item.Size().Height)
	objs[1].(*widget.Label).SetText(summaryMessage(node))
	objs[2].(*widget.Label).SetText(shortHash(node))
	objs[3].(*widget.Label).SetText(authorName(node))
	objs[4].(*widget.Label).SetText(commitedAt(node))
}

func summaryMessage(node *gogigu.Node) string {
	msg := strings.Split(node.Commit.Message, "\n")[0]
	if len(msg) > messageMaxLength {
		msg = msg[:messageMaxLength]
	}
	return msg
}

func shortHash(node *gogigu.Node) string {
	return node.ShortHash()
}

func authorName(node *gogigu.Node) string {
	return node.Commit.Author.Name
}

func commitedAt(node *gogigu.Node) string {
	return node.Commit.Author.When.Format(dateTimeFormat)
}

func (m *manager) buildCommitDetailView() fyne.CanvasObject {
	v := &commitDetailView{
		authorItemNameLabel:  widget.NewLabel(""),
		authorItemEmailLabel: widget.NewLabel(""),
		authorItemWhenLabel:  widget.NewLabel(""),
		hashItemLabel:        widget.NewLabel(""),
		parentsHashItemLabel: widget.NewLabel(""),
		messageItem:          widget.NewRichText(),
	}
	authorItemDetail := container.NewVBox(
		container.NewHBox(
			v.authorItemNameLabel,
			v.authorItemEmailLabel,
		),
		v.authorItemWhenLabel,
	)
	authorItem := widget.NewFormItem("Author", authorItemDetail)
	parentsHashItem := widget.NewFormItem("Parents", v.parentsHashItemLabel)
	hashItem := widget.NewFormItem("SHA", v.hashItemLabel)
	messageItem := widget.NewFormItem("", v.messageItem)
	v.Form = widget.NewForm(
		authorItem,
		hashItem,
		parentsHashItem,
		messageItem,
	)
	m.commitDetailView = v
	return container.NewVScroll(v.Form)
}

type commitDetailView struct {
	*widget.Form

	authorItemNameLabel  *widget.Label
	authorItemEmailLabel *widget.Label
	authorItemWhenLabel  *widget.Label
	hashItemLabel        *widget.Label
	parentsHashItemLabel *widget.Label
	messageItem          *widget.RichText
}

func (m *manager) updateCommitDetailView(id widget.ListItemID) {
	n := m.rm.Nodes[id]
	v := m.commitDetailView
	v.authorItemNameLabel.Text = n.Commit.Author.Name
	v.authorItemEmailLabel.Text = n.Commit.Author.Email
	v.authorItemWhenLabel.Text = n.Commit.Author.When.Format(dateTimeFormat)
	v.hashItemLabel.Text = n.Hash()
	v.parentsHashItemLabel.Text = m.parentsShortHashes(n)
	msgHead, msgTail := parseCommitMessage(n)
	msgHeadSegment := &widget.TextSegment{
		Style: widget.RichTextStyleSubHeading,
		Text:  msgHead,
	}
	msgTailSegment := &widget.TextSegment{
		Style: widget.RichTextStyleInline,
		Text:  msgTail,
	}
	v.messageItem.Segments = []widget.RichTextSegment{
		&widget.SeparatorSegment{},
		msgHeadSegment,
		msgTailSegment,
	}
	v.Form.Refresh()
}

func (m *manager) parentsShortHashes(n *gogigu.Node) string {
	ps := m.rm.Parents(n.Hash())
	hs := make([]string, len(ps))
	for i, p := range ps {
		hs[i] = p.ShortHash()
	}
	return strings.Join(hs, " ")
}

func parseCommitMessage(n *gogigu.Node) (string, string) {
	msgs := strings.SplitN(n.Commit.Message, "\n", 2)
	if len(msgs) > 1 {
		return msgs[0], msgs[1]
	}
	return msgs[0], ""
}

type sideMenuView struct {
	*widget.Tree
}

func (m *manager) buildSideMenuView() fyne.CanvasObject {
	v := &sideMenuView{}
	tree := widget.NewTreeWithStrings(map[string][]string{
		"":         {"Branches", "Remotes", "Tags"},
		"Branches": m.rm.BranchNames(),
		"Remotes":  m.rm.RemoteBranchNames(),
		"Tags":     m.rm.TagNames(),
	})
	v.Tree = tree
	m.sideMenuView = v
	return v.Tree
}
