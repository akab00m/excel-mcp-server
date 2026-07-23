package tools

import (
	"context"
	"fmt"

	z "github.com/Oudwins/zog"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/negokaz/excel-mcp-server/internal/excel"
	imcp "github.com/negokaz/excel-mcp-server/internal/mcp"
)

type ExcelCreateWorkbookArguments struct {
	FileAbsolutePath string `zog:"fileAbsolutePath"`
	SheetName        string `zog:"sheetName"`
}

var excelCreateWorkbookArgumentsSchema = z.Struct(z.Shape{
	"fileAbsolutePath": z.String().Test(AbsolutePathTest()).Required(),
	"sheetName":        z.String().Optional(),
})

func AddExcelCreateWorkbookTool(server *server.MCPServer) {
	server.AddTool(mcp.NewTool("excel_create_workbook",
		mcp.WithDescription("Create a new empty Excel workbook at the given path (fails if the file already exists)"),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(false),
		mcp.WithString("fileAbsolutePath",
			mcp.Required(),
			mcp.Description(FileAbsolutePathDescription),
		),
		mcp.WithString("sheetName",
			mcp.Description("Optional name for the first sheet (default sheet kept if omitted)"),
		),
	), handleCreateWorkbook)
}

func handleCreateWorkbook(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := ExcelCreateWorkbookArguments{}
	issues := excelCreateWorkbookArgumentsSchema.Parse(request.Params.Arguments, &args)
	if len(issues) != 0 {
		return imcp.NewToolResultZogIssueMap(issues), nil
	}
	return createWorkbook(args.FileAbsolutePath, args.SheetName)
}

func createWorkbook(fileAbsolutePath string, sheetName string) (*mcp.CallToolResult, error) {
	if err := excel.CreateFile(fileAbsolutePath, sheetName); err != nil {
		return nil, err
	}

	html := "<h2>Created Workbook</h2>\n<ul>\n"
	html += fmt.Sprintf("<li>path: %s</li>\n", fileAbsolutePath)
	if sheetName != "" {
		html += fmt.Sprintf("<li>sheet name: %s</li>\n", sheetName)
	}
	html += "</ul>\n"
	return mcp.NewToolResultText(html), nil
}
