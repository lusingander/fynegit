package ui

import (
	"fmt"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/lusingander/fynegit/internal/gogigu"
	"github.com/lusingander/fynegit/internal/graph"
	"github.com/lusingander/fynegit/internal/repository"
)

const (
	dateTimeFormat = "2006/01/02 15:04:05"

	graphMessageColumnWidth = 500.
	graphHashColumnWidth    = 80.
	graphAuthorColumnWidth  = 160.
)

var (
	defaultWindowSize = fyne.NewSize(1400, 800)
)

type manager struct {
	fyne.Window
	rm *repository.RepositoryManager

	*commitGraphView
	*commitDetailView
	*sideMenuView
	*patchSummaryView
}

func Start(w fyne.Window, rm *repository.RepositoryManager) {
	m := &manager{
		Window:           w,
		rm:               rm,
		commitGraphView:  nil,
		commitDetailView: nil,
		sideMenuView:     nil,
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

	commitInfoHs := container.NewHSplit(
		m.buildCommitDetailView(),
		m.buildPatchSummaryView(),
	)
	commitInfoHs.SetOffset(0.7)
	logVs := container.NewVSplit(
		m.buildCommitGraphView(),
		commitInfoHs,
	)
	logVs.SetOffset(0.6)
	repoHs := container.NewHSplit(
		m.buildSideMenuView(),
		logVs,
	)
	repoHs.SetOffset(0.15)
	return repoHs
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

type commitGraphView struct {
	*widget.List
}

func (m *manager) buildCommitGraphView() fyne.CanvasObject {
	v := &commitGraphView{}
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
	list.OnSelected = func(id widget.ListItemID) {
		m.updateCommitDetailView(id)
		m.updatePatchSummaryView(id)
	}
	v.List = list
	m.commitGraphView = v
	return list
}

func commitGraphItem(rm *repository.RepositoryManager) fyne.CanvasObject {
	graphAreaWidth := graph.CalcCommitGraphAreaWidth(rm)
	graphArea := widget.NewLabel("")
	refs := widget.NewLabel("")
	msg := widget.NewLabel("commit message")
	hash := widget.NewLabel("hash")
	author := widget.NewLabel("author")
	committedAt := widget.NewLabel("2006/01/02 15:04:05")
	var msgW, hashW, authorW float32 = graphMessageColumnWidth, graphHashColumnWidth, graphAuthorColumnWidth
	graphArea.Move(fyne.NewPos(0, 0))
	refs.Move(fyne.NewPos(graphArea.Position().X+graphAreaWidth, 0))
	msg.Move(fyne.NewPos(graphArea.Position().X+graphAreaWidth, 0))
	hash.Move(fyne.NewPos(msg.Position().X+msgW, 0))
	author.Move(fyne.NewPos(hash.Position().X+hashW, 0))
	committedAt.Move(fyne.NewPos(author.Position().X+authorW, 0))
	return container.NewWithoutLayout(
		graphArea,
		refs,
		msg,
		hash,
		author,
		committedAt,
	)
}

func updateCommitGraphItem(rm *repository.RepositoryManager, node *gogigu.Node, item fyne.CanvasObject) {
	objs := item.(*fyne.Container).Objects
	objs[0] = graph.CalcCommitGraphTreeRow(rm, node, item.Size().Height)
	refs, rw := calcCommitRefMarkers(rm, node, item.Size().Height)
	objs[1] = refs
	objs[2].(*widget.Label).SetText(summaryMessage(node, rw))
	objs[3].(*widget.Label).SetText(shortHash(node))
	objs[4].(*widget.Label).SetText(authorName(node))
	objs[5].(*widget.Label).SetText(commitedAt(node))
}

func calcCommitRefMarkers(rm *repository.RepositoryManager, node *gogigu.Node, h float32) (fyne.CanvasObject, float32) {
	refs := rm.AllRefs(node.Hash())
	if len(refs) == 0 {
		return container.NewWithoutLayout(), 0
	}

	markers, totalWidth := buildCommitRefMarkers(refs, h)

	var wBuf, hBuf float32 = theme.Padding(), 1
	limitWidth := graphMessageColumnWidth - wBuf*2
	if totalWidth > limitWidth {
		n := 1
		for len(refs) > 0 {
			refs = refs[:len(refs)-1]
			markers, totalWidth = buildCommitRefMarkers(refs, h)
			notice := fmt.Sprintf("+%d refs", n)
			noticeSize := textSize(notice)
			bg, fg := refsNoticeColor()
			rect := canvas.NewRectangle(bg)
			rectWidth := noticeSize.Width + wBuf*2
			rectHeight := noticeSize.Height + hBuf*2
			rect.Resize(fyne.NewSize(rectWidth, rectHeight))
			rect.Move(fyne.NewPos(wBuf+totalWidth, (h/2)-(rectHeight/2)))
			text := canvas.NewText(notice, fg)
			text.Move(fyne.NewPos(rect.Position().X+wBuf, rect.Position().Y+hBuf))
			totalWidth += rectWidth + wBuf*2
			if totalWidth < limitWidth {
				markers.(*fyne.Container).Objects = append(markers.(*fyne.Container).Objects, rect, text)
				break
			}
			n += 1
		}
	}
	left := graph.CalcCommitGraphAreaWidth(rm)
	markers.Move(fyne.NewPos(left, 0))
	return markers, totalWidth
}

func buildCommitRefMarkers(refs []*repository.Ref, h float32) (fyne.CanvasObject, float32) {
	var wBuf, hBuf float32 = theme.Padding(), 1
	var totalWidth float32 = 0
	objs := make([]fyne.CanvasObject, 0)
	for _, ref := range refs {
		bg, fg := refsColor(ref.RefType())
		rect := &canvas.Rectangle{
			FillColor:   bg,
			StrokeColor: fg,
			StrokeWidth: 1,
		}
		name := ref.Name()
		textSize := textSize(name)
		rectWidth := textSize.Width + wBuf*2
		rectHeight := textSize.Height + hBuf*2
		rect.Resize(fyne.NewSize(rectWidth, rectHeight))
		rect.Move(fyne.NewPos(wBuf+totalWidth, (h/2)-(rectHeight/2)))
		text := canvas.NewText(name, fg)
		text.Move(fyne.NewPos(rect.Position().X+wBuf, rect.Position().Y+hBuf))
		objs = append(objs, rect, text)
		totalWidth += rectWidth + wBuf*2
	}
	return container.NewWithoutLayout(objs...), totalWidth
}

func summaryMessage(node *gogigu.Node, refsWidth float32) string {
	msg := strings.Split(node.Commit.Message, "\n")[0]
	return dummyPaddingSpaces(refsWidth) + ellipsisText(msg, graphMessageColumnWidth-refsWidth)
}

func shortHash(node *gogigu.Node) string {
	return node.ShortHash()
}

func authorName(node *gogigu.Node) string {
	return ellipsisText(node.Commit.Author.Name, graphAuthorColumnWidth)
}

func commitedAt(node *gogigu.Node) string {
	return node.Commit.Author.When.Format(dateTimeFormat)
}

type commitDetailView struct {
	*container.Scroll
}

func (m *manager) buildCommitDetailView() fyne.CanvasObject {
	scroll := container.NewVScroll(widget.NewLabel(""))
	v := &commitDetailView{
		Scroll: scroll,
	}
	m.commitDetailView = v
	return v.Scroll
}

func (m *manager) updateCommitDetailView(id widget.ListItemID) {
	n := m.rm.Nodes[id]
	form := widget.NewForm()

	authorItemNameLabel := widget.NewLabel(n.Commit.Author.Name)
	authorItemEmailLabel := widget.NewLabel(formatEmail(n.Commit.Author.Email))
	authorItemWhenLabel := widget.NewLabel(n.Commit.Author.When.Format(dateTimeFormat))
	authorItemDetail := container.NewHBox(
		authorItemNameLabel,
		authorItemEmailLabel,
		authorItemWhenLabel,
	)
	form.Append("Author", authorItemDetail)

	if showCommiter(n) {
		committerItemNameLabel := widget.NewLabel(n.Commit.Committer.Name)
		committerItemEmailLabel := widget.NewLabel(formatEmail(n.Commit.Committer.Email))
		committerItemWhenLabel := widget.NewLabel(n.Commit.Committer.When.Format(dateTimeFormat))
		committerItemDetail := container.NewHBox(
			committerItemNameLabel,
			committerItemEmailLabel,
			committerItemWhenLabel,
		)
		form.Append("Committer", committerItemDetail)
	}

	hashItemLabel := widget.NewLabel(n.Hash())
	form.Append("SHA", hashItemLabel)

	parentsHashItemLabel := widget.NewLabel(m.parentsShortHashes(n))
	form.Append("Parents", parentsHashItemLabel)

	if refs := m.rm.AllRefs(n.Hash()); len(refs) > 0 {
		dummy := widget.NewLabel("")
		dh := dummy.Size().Height
		markers, _ := buildCommitRefMarkers(refs, dh)
		refsItem := widget.NewFormItem("Refs", container.NewWithoutLayout(dummy, markers))
		markers.Move(fyne.NewPos(0, (refsItem.Widget.Size().Height/2)-(dh/2)))
		form.AppendItem(refsItem)
	}

	messageItemRichText := widget.NewRichText()
	messageItemRichText.Wrapping = fyne.TextWrapWord
	msgHead, msgTail := parseCommitMessage(n)
	msgHeadSegment := &widget.TextSegment{
		Style: widget.RichTextStyleSubHeading,
		Text:  msgHead,
	}
	msgTailSegment := &widget.TextSegment{
		Style: widget.RichTextStyleInline,
		Text:  msgTail,
	}
	messageItemRichText.Segments = []widget.RichTextSegment{
		&widget.SeparatorSegment{},
		msgHeadSegment,
		msgTailSegment,
	}
	form.Append("", messageItemRichText)

	v := m.commitDetailView
	v.Scroll.Content = form
	v.Scroll.Refresh()
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

func formatEmail(email string) string {
	return fmt.Sprintf("<%s>", email)
}

func showCommiter(n *gogigu.Node) bool {
	a := n.Commit.Author
	c := n.Commit.Committer
	return a.Name != c.Name || a.Email != c.Email || !a.When.Equal(c.When)
}

type sideMenuView struct {
	*widget.Tree
}

func (m *manager) buildSideMenuView() fyne.CanvasObject {
	v := &sideMenuView{}
	tree := widget.NewTreeWithStrings(map[string][]string{
		"":                {"Local Branches", "Remote Branches", "Tags"},
		"Local Branches":  m.rm.BranchNames(),
		"Remote Branches": m.rm.RemoteBranchNames(),
		"Tags":            m.rm.SortedTagNames(),
	})
	tree.OnSelected = m.selectRefRow
	v.Tree = tree
	m.sideMenuView = v
	return v.Tree
}

func (m *manager) selectRefRow(name string) {
	list := m.commitGraphView.List
	if list == nil {
		return
	}
	ref := m.rm.FromRefName(name)
	if ref == nil {
		return
	}
	refNode := m.rm.Node(ref.TargetHash())
	if refNode == nil {
		return
	}
	list.Select(refNode.PosY())
}

func ellipsisText(src string, maxWidth float32) string {
	buf := textWidth("__")
	if textWidth(src) < maxWidth-buf {
		return src
	}

	tail := "..."
	tailBuf := textWidth(tail) + buf
	if maxWidth < tailBuf {
		panic(fmt.Errorf("maxWidth %v must be less than %v", maxWidth, tailBuf))
	}

	cut := cutText(src, maxWidth-tailBuf, buf, 0, 6)
	return cut + tail
}

type patchSummaryView struct {
	*container.Scroll
}

func (m *manager) buildPatchSummaryView() fyne.CanvasObject {
	v := &patchSummaryView{}
	scroll := container.NewScroll(widget.NewLabel(""))
	v.Scroll = scroll
	m.patchSummaryView = v
	return v.Scroll
}

func (m *manager) updatePatchSummaryView(id widget.ListItemID) {
	n := m.rm.Nodes[id]
	v := m.patchSummaryView
	details, err := m.rm.PatchFileDetails(n)
	if err != nil {
		v.Scroll.Content = widget.NewLabel("")
		v.Scroll.Refresh()
		return
	}
	rows := make([]fyne.CanvasObject, len(details))
	for i, d := range details {
		rows[i] = buildChangeDetailLine(d)
	}
	v.Scroll.Content = container.NewVBox(rows...)
	v.Scroll.Refresh()
}

func buildChangeDetailLine(d *repository.PatchFileDetail) fyne.CanvasObject {
	var icon fyne.Resource
	switch d.ChangeType() {
	case repository.Modify:
		icon = theme.MoreHorizontalIcon()
	case repository.Insert:
		icon = theme.ContentAddIcon()
	case repository.Delete:
		icon = theme.ContentRemoveIcon()
	case repository.Move:
		icon = theme.NavigateNextIcon()
	}
	return container.NewHBox(
		widget.NewIcon(icon),
		widget.NewLabel(d.Name()),
	)
}
