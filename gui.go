package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Problem struct {
	ID         string
	Title      string
	Difficulty string
	Premium    bool
	Tags       []string
	Frequency  int
}

var (
	allProblems      []Problem
	filteredProblems []Problem
	currentFilter    = "全部"
	currentSort      = "题号"
	sortAsc          = true
	tagList          = []string{
		"数组", "字符串", "哈希表", "数学", "动态规划", "排序",
		"贪心", "深度优先搜索", "二分查找", "数据库", "树",
		"位运算", "矩阵", "广度优先搜索", "双指针", "前缀和",
		"堆（优先队列）", "二叉树", "模拟", "栈", "图", "计数",
		"滑动窗口", "设计", "枚举", "回溯", "链表", "并查集",
		"数论", "有序集合", "线段树", "单调栈", "分治", "字典树",
		"递归", "组合数学", "状态压缩", "队列", "二叉搜索树",
		"几何", "记忆化搜索", "树状数组", "哈希函数", "拓扑排序",
		"最短路", "字符串匹配", "滚动哈希", "博弈", "数据流",
		"交互", "单调队列", "脑筋急转弯", "双向链表", "归并排序",
		"随机化", "快速选择", "计数排序", "迭代器", "概率与统计",
		"多线程", "扫描线", "后缀数组", "桶排序", "最小生成树",
		"Shell", "水塘抽样", "强连通分量", "欧拉回路", "基数排序",
		"双连通分量", "拒绝采样",
	}
)

func parseLine(line string) *Problem {
	re := regexp.MustCompile(`\s{2,}`)
	parts := re.Split(line, -1)
	var fields []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			fields = append(fields, p)
		}
	}
	if len(fields) < 4 {
		return nil
	}
	if len(fields) >= 2 {
		switch fields[0] {
		case "LCP", "LCR", "LCS", "面试题":
			fields = append([]string{fields[0] + " " + fields[1]}, fields[2:]...)
		}
	}
	id := fields[0]
	if id == "题号" {
		return nil
	}
	var title, diff, prem, tagsStr string
	title = fields[1]
	diff = fields[2]
	freqIdx := len(fields) - 1
	if freqIdx < 3 {
		return nil
	}
	if fields[3] == "会员" {
		prem = fields[3]
		if freqIdx >= 5 {
			tagsStr = fields[4]
		} else {
			tagsStr = ""
		}
	} else {
		prem = ""
		if freqIdx == 3 {
			tagsStr = ""
		} else {
			tagsStr = fields[3]
		}
	}
	frequency, _ := strconv.Atoi(fields[freqIdx])
	var tags []string
	if tagsStr != "" {
		for _, t := range strings.Split(tagsStr, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tags = append(tags, t)
			}
		}
	}
	return &Problem{
		ID:         id,
		Title:      title,
		Difficulty: diff,
		Premium:    prem == "会员",
		Tags:       tags,
		Frequency:  frequency,
	}
}

func loadProblems(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	allProblems = nil
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum <= 3 {
			continue
		}
		p := parseLine(scanner.Text())
		if p != nil {
			allProblems = append(allProblems, *p)
		}
	}
	filteredProblems = make([]Problem, len(allProblems))
	copy(filteredProblems, allProblems)
	return scanner.Err()
}

func applyFilter(filter string) {
	currentFilter = filter
	filteredProblems = nil
	for _, p := range allProblems {
		if matchFilter(p, currentFilter) {
			filteredProblems = append(filteredProblems, p)
		}
	}
	applySort(currentSort, sortAsc)
}

func idSortKey(id string) string {
	if n, err := strconv.Atoi(id); err == nil {
		return fmt.Sprintf("0_%010d", n)
	}
	if strings.HasPrefix(id, "LCP") {
		rest := strings.TrimSpace(strings.TrimPrefix(id, "LCP"))
		if n, err := strconv.Atoi(rest); err == nil {
			return fmt.Sprintf("1_%010d", n)
		}
	}
	if strings.HasPrefix(id, "LCR") {
		rest := strings.TrimSpace(strings.TrimPrefix(id, "LCR"))
		if n, err := strconv.Atoi(rest); err == nil {
			return fmt.Sprintf("2_%010d", n)
		}
	}
	if strings.HasPrefix(id, "LCS") {
		rest := strings.TrimSpace(strings.TrimPrefix(id, "LCS"))
		if n, err := strconv.Atoi(rest); err == nil {
			return fmt.Sprintf("3_%010d", n)
		}
	}
	if strings.HasPrefix(id, "面试题") {
		rest := strings.TrimSpace(strings.TrimPrefix(id, "面试题"))
		return fmt.Sprintf("4_%s", rest)
	}
	return fmt.Sprintf("9_%s", id)
}

func applySort(sortBy string, asc bool) {
	currentSort = sortBy
	sortAsc = asc
	sort.Slice(filteredProblems, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "频率":
			less = filteredProblems[i].Frequency < filteredProblems[j].Frequency
		default:
			less = idSortKey(filteredProblems[i].ID) < idSortKey(filteredProblems[j].ID)
		}
		if !asc {
			less = !less
		}
		return less
	})
}

func applySearch(keyword string) {
	if keyword == "" {
		applyFilter(currentFilter)
		return
	}
	keyword = strings.ToLower(keyword)
	filteredProblems = nil
	for _, p := range allProblems {
		if !matchFilter(p, currentFilter) {
			continue
		}
		if strings.Contains(strings.ToLower(p.ID), keyword) || strings.Contains(strings.ToLower(p.Title), keyword) || strings.Contains(strings.ToLower(strings.Join(p.Tags, ",")), keyword) {
			filteredProblems = append(filteredProblems, p)
		}
	}
	applySort(currentSort, sortAsc)
}

func matchFilter(p Problem, filter string) bool {
	switch filter {
	case "全部":
		return true
	case "会员":
		return p.Premium
	case "LCP":
		return strings.HasPrefix(p.ID, "LCP")
	case "LCR":
		return strings.HasPrefix(p.ID, "LCR")
	case "LCS":
		return strings.HasPrefix(p.ID, "LCS")
	case "面试题":
		return strings.HasPrefix(p.ID, "面试题")
	case "普通题":
		return !strings.HasPrefix(p.ID, "LCP") && !strings.HasPrefix(p.ID, "LCR") && !strings.HasPrefix(p.ID, "LCS") && !strings.HasPrefix(p.ID, "面试题")
	default:
		for _, t := range p.Tags {
			if t == filter {
				return true
			}
		}
		return false
	}
}

func main() {
	if err := loadProblems("data/example.txt"); err != nil {
		fmt.Println("加载失败:", err)
		return
	}
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("AlgoTop")
	w.Resize(fyne.NewSize(1500, 800))
	statusLabel := widget.NewLabel(fmt.Sprintf("共 %d 道题目 | 标签 %d 个", len(allProblems), len(tagList)))
	var table *widget.Table
	table = widget.NewTable(
		func() (int, int) { return len(filteredProblems) + 1, 7 },
		func() fyne.CanvasObject {
			return widget.NewLabel("placeholder text here")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id.Row == 0 {
				label.TextStyle = fyne.TextStyle{Bold: true}
				switch id.Col {
				case 0:
					label.SetText("序号")
				case 1:
					label.SetText("题号")
				case 2:
					label.SetText("题目名称")
				case 3:
					label.SetText("难度")
				case 4:
					label.SetText("会员")
				case 5:
					label.SetText("频率")
				case 6:
					label.SetText("标签")
				}
			} else {
				label.TextStyle = fyne.TextStyle{}
				p := filteredProblems[id.Row-1]
				switch id.Col {
				case 0:
					label.SetText(fmt.Sprintf("%d", id.Row))
				case 1:
					label.SetText(p.ID)
				case 2:
					label.SetText(p.Title)
				case 3:
					label.SetText(p.Difficulty)
				case 4:
					if p.Premium {
						label.SetText("是")
					} else {
						label.SetText("否")
					}
				case 5:
					label.SetText(fmt.Sprintf("%d", p.Frequency))
				case 6:
					label.SetText(strings.Join(p.Tags, ","))
				}
			}
		},
	)
	table.SetColumnWidth(0, 60)
	table.SetColumnWidth(1, 120)
	table.SetColumnWidth(2, 280)
	table.SetColumnWidth(3, 60)
	table.SetColumnWidth(4, 50)
	table.SetColumnWidth(5, 100)
	table.SetColumnWidth(6, 300)
	filterOpts := []string{"全部", "普通题", "LCP", "LCR", "LCS", "面试题", "会员"}
	filterOpts = append(filterOpts, tagList...)
	filterSelect := widget.NewSelect(filterOpts, func(s string) {
		applyFilter(s)
		statusLabel.SetText(fmt.Sprintf("筛选: %s | %d 道", currentFilter, len(filteredProblems)))
		table.Refresh()
	})
	filterSelect.SetSelected("全部")
	sortOpts := []string{"题号", "频率"}
	sortSelect := widget.NewSelect(sortOpts, func(s string) {
		applySort(s, sortAsc)
		table.Refresh()
	})
	sortSelect.SetSelected("题号")
	orderOpts := []string{"升序", "降序"}
	orderSelect := widget.NewSelect(orderOpts, func(s string) {
		sortAsc = s == "升序"
		applySort(currentSort, sortAsc)
		table.Refresh()
	})
	orderSelect.SetSelected("升序")
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("题号/名称/标签")
	searchEntry.OnChanged = func(s string) {
		applySearch(s)
		statusLabel.SetText(fmt.Sprintf("搜索: %d 道", len(filteredProblems)))
		table.Refresh()
	}
	searchBox := container.New(layout.NewGridWrapLayout(fyne.NewSize(300, 36)), searchEntry)
	toolbar := container.NewHBox(
		widget.NewLabel("筛选"), filterSelect,
		widget.NewLabel("排序"), sortSelect, orderSelect,
		widget.NewLabel("搜索"), searchBox,
	)
	content := container.NewBorder(
		toolbar,
		statusLabel,
		nil, nil,
		table,
	)
	w.SetContent(content)
	w.ShowAndRun()
}
