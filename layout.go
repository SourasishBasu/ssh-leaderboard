package main

// This example demonstrates various Lip Gloss style and layout features.

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/gamut"
	"github.com/sourasishbasu/test/db"
	"golang.org/x/term"
)

const (
	// In real life situations we'd adjust the document to fit the width we've
	// detected. In the case of this example we're hardcoding the width, and
	// later using the detected width only to truncate in order to avoid jaggy
	// wrapping.
	width = 96
)

// Style definitions.
var (

	// General.

	normal    = lipgloss.Color("#EEEEEE")
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7A2782"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#faef1d"}
	blends    = gamut.Blends(lipgloss.Color("#F25D94"), lipgloss.Color("#EDFF82"), 50)

	base = lipgloss.NewStyle().Foreground(normal)

	divider = lipgloss.NewStyle().
		SetString("•").
		Padding(0, 1).
		Foreground(subtle).
		String()

	//url = lipgloss.NewStyle().Foreground(lipgloss.Color("#D12881")).Render
	heading = lipgloss.NewStyle().Foreground(special).Render

	// Tabs.

	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	tab = lipgloss.NewStyle().
		Border(tabBorder, true).
		BorderForeground(highlight).
		Padding(0, 1)

	activeTab = tab.Border(activeTabBorder, true)

	tabGap = tab.
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)

	// Title.

	titleStyle = lipgloss.NewStyle().
			MarginLeft(1).
			MarginRight(5).
			Padding(0, 1).
			Italic(true).
			Foreground(lipgloss.Color("#FFF7DB")).
			SetString("Lip Gloss")

	descStyle = base.MarginTop(1)

	infoStyle = base.
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(subtle)

	// Status Bar.

	statusNugget = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			Background(lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#353533"})

	statusStyle = lipgloss.NewStyle().
			Inherit(statusBarStyle).
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#D12881")).
			Padding(0, 1).
			MarginRight(1)

	statusText = lipgloss.NewStyle().Inherit(statusBarStyle)

	fishCakeStyle = statusNugget.Background(lipgloss.Color("#7A2782"))

	// Page.

	docStyle = lipgloss.NewStyle().Padding(1, 2, 1, 2)
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(highlight).
	Margin(0, 25).
	Bold(true)

const (
	host = "localhost"
	port = "23234"
)

type model struct {
	table table.Model
	rows  []table.Row
}

// Initialize env at package level
func init() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file")
	}
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	_, _, active := s.Pty()
	if !active {
		fmt.Println("No active terminal, skipping")
		return nil, nil
	}

	// Create our program with the initial model
	m := initialModel()

	return m, []tea.ProgramOption{
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	}
}

func startServer() {
	// Create the SSH server
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}

func getMembers() []table.Row {
	connStr := os.Getenv("DATABASE_ENDPOINT")
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	queries := db.New(conn)
	members, err := queries.ListMembers(context.Background())
	if err != nil {
		fmt.Println("Error fetching members:", err)
		os.Exit(1)
	}

	var rows []table.Row
	for _, member := range members {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", member.Rank),
			member.Name,
			fmt.Sprintf("%d", member.Scores),
		})
	}
	return rows
}

type tickMsg time.Time

func tickEvery() tea.Cmd {
	return tea.Every(10*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func initialModel() model {

	t := table.New(
		table.WithColumns([]table.Column{
			{Title: "PLACE", Width: 7},
			{Title: "NAME", Width: 15},
			{Title: "SCORES", Width: 9},
		}),
		table.WithHeight(20),
		table.WithFocused(true),
	)
	//t.Focus()

	// Set the styles
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(highlight).
		BorderTop(false).BorderBottom(true).BorderLeft(false).BorderRight(false).
		Bold(true).
		Align(lipgloss.Center, lipgloss.Center)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("233")).
		Background(special).
		Bold(false)
	s.Cell = s.Cell.
		Align(lipgloss.Center, lipgloss.Center)
	t.SetStyles(s)

	// Get initial rows
	initialRows := getMembers()

	return model{
		table: t,
		rows:  initialRows,
	}
}

func (m model) Init() tea.Cmd { return tickEvery() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tickMsg:
		// Fetch fresh data
		newRows := getMembers()
		m.rows = newRows
		m.table.SetRows(newRows)
		return m, tickEvery()
	case tea.KeyMsg:
		switch msg.String() {
		case "b":
			m.table.Focus()
			return m, cmd
		case "esc":
			m.table.Blur()
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Printf("%s is ranked %s", m.table.SelectedRow()[1], m.table.SelectedRow()[0])
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {

	physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))
	doc := strings.Builder{}

	// Tabs
	{
		row := lipgloss.JoinHorizontal(
			lipgloss.Top,
			activeTab.Render(rainbow(lipgloss.NewStyle(), "Welcome to Lost Messages", blends)),
			tab.Render(rainbow(lipgloss.NewStyle(), "Prizes upto 16k INR!!", blends)),
			tab.Render(rainbow(lipgloss.NewStyle(), "Goodies, Stickers and more", blends)),
		)
		gap := tabGap.Render(strings.Repeat(" ", max(0, width-lipgloss.Width(row)-2)))
		row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)
		doc.WriteString(row + "\n\n")
	}

	// Title
	{
		var (
			colors = colorGrid(1, 5)
			title  strings.Builder
		)

		for i, v := range colors {
			const offset = 2
			c := lipgloss.Color(v[0])
			fmt.Fprint(&title, titleStyle.MarginLeft(i*offset).Background(c))
			if i < len(colors)-1 {
				title.WriteRune('\n')
			}
		}

		desc := lipgloss.JoinVertical(lipgloss.Left,
			descStyle.Render("🔴 LIVE Leaderboard - "+rainbow(lipgloss.NewStyle(), "Lost Messages by KIITFEST 8.0", blends)),
			infoStyle.Render("Visit"+divider+heading("https://lostmessages.mlsakiit.com")+divider+"to play"),
		)

		row := lipgloss.JoinHorizontal(lipgloss.Top, title.String(), desc)
		doc.WriteString(row + "\n")
	}

	// Table
	{
		m.table.SetRows(m.rows)
		doc.WriteString(baseStyle.Render(m.table.View()) + "\n\n")
	}

	// Status bar
	{
		w := lipgloss.Width

		statusKey := statusStyle.Render("STATUS")
		fishCake := fishCakeStyle.Render("🕵️ Lost Messages")
		statusVal := statusText.
			Width(width - w(statusKey) - w(fishCake)).
			Render("Refreshing every 5 secs...")

		bar := lipgloss.JoinHorizontal(lipgloss.Top,
			statusKey,
			statusVal,
			fishCake,
		)

		doc.WriteString(statusBarStyle.Width(width).Render(bar))
	}

	if physicalWidth > 0 {
		docStyle = docStyle.MaxWidth(physicalWidth)
	}

	return doc.String()

	// Okay, let's print it
	//fmt.Println(docStyle.Render(doc.String()))
}

func main() {
	startServer()
	//p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	//if _, err := p.Run(); err != nil {
	//	fmt.Println("Error running program:", err)
	//	os.Exit(1)
	//}
}

func colorGrid(xSteps, ySteps int) [][]string {
	x0y0, _ := colorful.Hex("#F25D94")
	x1y0, _ := colorful.Hex("#EDFF82")
	x0y1, _ := colorful.Hex("#643AFF")
	x1y1, _ := colorful.Hex("#14F9D5")

	x0 := make([]colorful.Color, ySteps)
	for i := range x0 {
		x0[i] = x0y0.BlendLuv(x0y1, float64(i)/float64(ySteps))
	}

	x1 := make([]colorful.Color, ySteps)
	for i := range x1 {
		x1[i] = x1y0.BlendLuv(x1y1, float64(i)/float64(ySteps))
	}

	grid := make([][]string, ySteps)
	for x := 0; x < ySteps; x++ {
		y0 := x0[x]
		grid[x] = make([]string, xSteps)
		for y := 0; y < xSteps; y++ {
			grid[x][y] = y0.BlendLuv(x1[x], float64(y)/float64(xSteps)).Hex()
		}
	}

	return grid
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func rainbow(base lipgloss.Style, s string, colors []color.Color) string {
	var str string
	for i, ss := range s {
		color, _ := colorful.MakeColor(colors[i%len(colors)])
		str = str + base.Foreground(lipgloss.Color(color.Hex())).Render(string(ss))
	}
	return str
}
