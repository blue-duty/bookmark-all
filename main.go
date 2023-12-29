package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/net/html"
)

type Bookmarks struct {
	Title        string       `json:"Title"`
	URL          string       `json:"URL"`
	AddDate      string       `json:"AddDate"`
	Icon         string       `json:"Icon"`
	IsDir        bool         `json:"IsDir"`
	LastModified string       `json:"LastModified"`
	Children     []*Bookmarks `json:"Children"`
}

func bookmark2json(fName string) (b *Bookmarks, err error) {
	file, err := os.Open(fName)
	if err != nil {
		return
	}
	defer file.Close()
	doc, err := html.Parse(file)
	if err != nil {
		return
	}

	var f func(*html.Node, *Bookmarks) *Bookmarks
	f = func(n *html.Node, b *Bookmarks) *Bookmarks {
		if n.Type == html.ElementNode && n.Data == "a" {
			var bi = &Bookmarks{IsDir: false}
			for _, a := range n.Attr {
				if a.Key == "href" {
					bi.URL = a.Val
				}
				if a.Key == "icon" {
					bi.Icon = a.Val
				}
				if a.Key == "add_date" {
					bi.AddDate = a.Val
				}
			}
			if n.FirstChild != nil {
				bi.Title = n.FirstChild.Data
			}
			b.Children = append(b.Children, bi)
			return b
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "h3" {
				var text string
				if c.FirstChild != nil {
					text = c.FirstChild.Data
				}

				dir := &Bookmarks{Title: text, IsDir: true}
				for _, a := range c.Attr {
					if a.Key == "add_date" {
						dir.AddDate = a.Val
					} else if a.Key == "last_modified" {
						dir.LastModified = a.Val
					}
				}

				b.Children = append(b.Children, dir)
				b = dir
			} else if c.Type != html.TextNode && c.Type != html.DoctypeNode && c.Type != html.CommentNode {
				f(c, b)
			}
		}
		return b
	}
	b = f(doc, &Bookmarks{Title: "Bookmarks", IsDir: true})
	return
}

// 比较两个书签文件，汇总成一个
// 比较两个书签文件，汇总成一个
func compareBookmark(b1, b2 *Bookmarks) (b *Bookmarks) {
	b = &Bookmarks{Title: "Merged Bookmarks", IsDir: true} // 创建合并后的根节点
	if b1.Title == b2.Title {
		b.Title = b1.Title
	}
	for _, v := range b1.Children {
		found := false
		if v.IsDir {
			for _, v2 := range b2.Children {
				if v2.IsDir && v.Title == v2.Title {
					found = true
					b.Children = append(b.Children, compareBookmark(v, v2))
					break
				}
			}
		} else {
			for _, v2 := range b2.Children {
				if !v2.IsDir && v.Title == v2.Title {
					found = true
					b.Children = append(b.Children, v)
					break
				}
			}
		}
		// 未匹配到的文件夹或书签也需要添加到结果中
		if !found {
			b.Children = append(b.Children, v)
		}
	}
	// 处理在 b2 中而不在 b1 中的文件夹和书签
	for _, v2 := range b2.Children {
		found := false
		for _, v := range b1.Children {
			if v.IsDir && v2.IsDir && v.Title == v2.Title {
				found = true
				break
			} else if !v.IsDir && !v2.IsDir && v.Title == v2.Title {
				found = true
				break
			}
		}
		// 未匹配到的文件夹或书签也需要添加到结果中
		if !found {
			b.Children = append(b.Children, v2)
		}
	}
	return b
}

// 将 Bookmarks 结构转换为 HTML 书签
func bookmarksToHTML(b *Bookmarks) string {
	var result string

	if b.IsDir {
		// 处理文件夹
		result += fmt.Sprintf("<DL><p>\n    <DT><H3>%s</H3>\n", escapeHTML(b.Title))
		for _, child := range b.Children {
			result += bookmarksToHTML(child)
		}
		result += "</DL><p>\n"
	} else {
		// 处理书签
		result += fmt.Sprintf("<DT><A HREF=\"%s\" ADD_DATE=\"%s\" ICON=\"%s\">%s</A>\n",
			escapeHTML(b.URL), b.AddDate, b.Icon, escapeHTML(b.Title))
	}

	return result
}

// 将字符串进行 HTML 转义
func escapeHTML(s string) string {
	return strings.ReplaceAll(s, "&", "&amp;")
}

func main() {
	edge, _ := bookmark2json("edge.html")
	chrome, _ := bookmark2json("chrome.html")
	firefox, _ := bookmark2json("firefox.html")

	b := compareBookmark(edge, chrome)
	b = compareBookmark(b, firefox)

	// 将 Bookmarks 结构转换为 HTML 书签
	var header = `<!DOCTYPE NETSCAPE-Bookmark-file-1>
<!-- This is an automatically generated file.
	 It will be read and overwritten.
	 DO NOT EDIT! -->
	 <META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
	 <TITLE>Bookmarks</TITLE>
	 <H1>Bookmarks</H1>
	 <DL><p>
`
	htmlBookmarks := bookmarksToHTML(b)

	htmlBookmarks = header + htmlBookmarks + "</DL><p>\n"

	// 去掉9 , 10 , 11 行
	htmlBookmarks = strings.Replace(htmlBookmarks, "<DT><H3>Bookmarks</H3>\n", "", 1)
	htmlBookmarks = strings.Replace(htmlBookmarks, "<DL><p>\n", "", 2)
	//保存在文件中
	f, err := os.Create("bookmarks.html")
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	defer f.Close()
	_, err = f.WriteString(htmlBookmarks)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
