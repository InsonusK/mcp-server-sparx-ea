// Command modelgen builds the ArchiMate model of this MCP server itself and
// writes it to docs/mcp-server-sparx-ea.arch.xml, an EA "Import Package from
// XMI" file. It uses internal/service/sparx — the same service the MCP tools
// call — so the model is produced exactly as ea_new_model / ea_create_* would,
// diagrams included.
//
//	go run ./tools/modelgen            # writes docs/mcp-server-sparx-ea.arch.xml
//	go run ./tools/modelgen -o out.xml
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
)

func main() {
	out := flag.String("o", "docs/mcp-server-sparx-ea.arch.xml", "path to write the model to")
	flag.Parse()

	m, err := sparx.NewModel("Sparx EA MCP Server")
	if err != nil {
		log.Fatal(err)
	}
	b := &builder{m: m}

	b.capabilities()
	b.roadmap()
	b.diagrams()

	if b.err != nil {
		log.Fatal(b.err)
	}
	if err := m.Save(*out); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %s (%d elements, %d relationships, %d diagrams)\n", *out, b.elems, b.rels, b.diags)
}

type builder struct {
	m                  *sparx.Service
	err                error
	elems, rels, diags int
}

func (b *builder) pkg(parent, name string) string {
	if b.err != nil {
		return ""
	}
	p, err := b.m.CreatePackage(parent, name)
	if err != nil {
		b.err = fmt.Errorf("package %s/%s: %w", parent, name, err)
		return ""
	}
	return p.Path
}

func (b *builder) el(parent, typ, name, note string) string {
	if b.err != nil {
		return ""
	}
	e, err := b.m.CreateElement(parent, typ, name, note)
	if err != nil {
		b.err = fmt.Errorf("element %s %q: %w", typ, name, err)
		return ""
	}
	b.elems++
	return e.Path
}

func (b *builder) rel(typ, src, tgt, name string) {
	if b.err != nil || src == "" || tgt == "" {
		return
	}
	if _, err := b.m.CreateRelationship(src, tgt, typ, name, ""); err != nil {
		b.err = fmt.Errorf("relationship %s %s -> %s: %w", typ, src, tgt, err)
		return
	}
	b.rels++
}

func (b *builder) dgm(parent, name, layer string) string {
	if b.err != nil {
		return ""
	}
	g, err := b.m.CreateDiagram(parent, name, layer)
	if err != nil {
		b.err = fmt.Errorf("diagram %s/%s: %w", parent, name, err)
		return ""
	}
	b.diags++
	return g.Path
}

func (b *builder) place(diagram, element string, left, top int) {
	if b.err != nil || diagram == "" || element == "" {
		return
	}
	at := sparx.Rect{Left: left, Top: top, Right: left + boxW, Bottom: top + boxH}
	if err := b.m.AddToDiagram(diagram, element, at); err != nil {
		b.err = fmt.Errorf("place %s on %s: %w", element, diagram, err)
	}
}

// ---------- capabilities ----------

type group struct {
	name  string
	note  string
	tools string
}

// groups is one entry per ArchiMate ApplicationFunction / ApplicationService
// pair — a group of ea_* tools the server exposes.
var groups = []group{
	{"Чтение модели", "Навигация по дереву и чтение одного элемента, пакета или диаграммы.",
		"ea_model_tree(file); ea_element(file, ref); ea_package(file, ref); ea_diagram(file, ref)"},
	{"Словарь типов ArchiMate", "Список принимаемых сервером типов элементов и связей ArchiMate 3.2.",
		"ea_archimate_types()"},
	{"Управление элементами", "Создание, переименование, смена документации и перенос, удаление элементов ArchiMate.",
		"ea_create_element; ea_update_element; ea_delete_element"},
	{"Управление связями", "Создание связей ArchiMate с проверкой правил и удаление по id или паре концов.",
		"ea_create_relationship; ea_delete_relationship"},
	{"Управление пакетами", "Создание, переименование и перенос, глубокое копирование, удаление пакетов.",
		"ea_create_package; ea_update_package; ea_copy_package; ea_delete_package"},
	{"Диаграммы", "Создание диаграммы и размещение / перемещение / снятие элементов на ней (координаты EA).",
		"ea_create_diagram(file, output, parent, name, layer?); ea_place_on_diagram; ea_move_on_diagram; ea_remove_from_diagram"},
	{"Жизненный цикл модели", "Создание модели с нуля (ArchiMate3-профиль встроен), добавление и переименование корневого пакета.",
		"ea_new_model; ea_create_root_package; ea_set_root_name"},
	{"Отправка bug report", "Запланировано. Отправка структурированного отчёта об ошибке сопровождающему (номер ошибки, контекст, входной файл).",
		"ea_bug_report(code, summary, context?)"},
	{"Запрос новых типов ArchiMate", "Запланировано. Запрос на расширение словаря типов при неизвестном стереотипе или нехватке типов агенту.",
		"ea_request_type(name, kind, rationale)"},
}

// deliveredBy names the work package that realizes each capability service
// (roadmap step). "" means not delivered by a concrete step yet.
var deliveredBy = map[string]string{
	"Чтение модели":                "MVP",
	"Словарь типов ArchiMate":      "MVP",
	"Управление элементами":        "MVP",
	"Управление связями":           "MVP",
	"Управление пакетами":          "Доп функции",
	"Диаграммы":                    "Доп функции",
	"Жизненный цикл модели":        "Доп функции",
	"Отправка bug report":          "Отправка bug report",
	"Запрос новых типов ArchiMate": "Запрос добавления типов",
}

const (
	compName = "mcp-server-sparx-ea"
	fnPath   = "Sparx EA MCP Server/Возможности сервиса/Функции/"
	svcPath  = "Sparx EA MCP Server/Возможности сервиса/Сервисы/"
	compPath = "Sparx EA MCP Server/Возможности сервиса/Приложение/" + compName
	wpPath   = "Sparx EA MCP Server/План развития/Работы/Работа: "
	plPath   = "Sparx EA MCP Server/План развития/"
)

func (b *builder) capabilities() {
	root := b.pkg("Sparx EA MCP Server", "Возможности сервиса")
	appPkg := b.pkg(root, "Приложение")
	fnPkg := b.pkg(root, "Функции")
	svcPkg := b.pkg(root, "Сервисы")

	comp := b.el(appPkg, "ArchiMate.ApplicationComponent", compName,
		"MCP-сервер (stdio) над экспортом Sparx EA в XMI 2.1. Читает и редактирует копию модели, "+
			"отдавая новый файл для повторного импорта в EA. Чистый Go, один статический бинарник, без SQL. "+
			"Слои: client/eaxmi (кодек XMI 2.1), internal/service/sparx (словарь и валидация ArchiMate 3.2), "+
			"internal/mcpserver (инструменты).")

	for _, g := range groups {
		fn := b.el(fnPkg, "ArchiMate.ApplicationFunction", g.name, g.note+"\n\nИнструменты: "+g.tools)
		svc := b.el(svcPkg, "ArchiMate.ApplicationService", g.name, g.note)
		b.rel("ArchiMate.Assignment", comp, fn, "")
		b.rel("ArchiMate.Realization", fn, svc, "")
	}
}

// ---------- roadmap ----------

type step struct {
	key  string // work-package suffix / plateau id (without the leading number)
	name string // plateau display name ("1. MVP")
	note string
}

var steps = []step{
	{"MVP", "1. MVP",
		"Готово. Базовые read/write инструменты над экспортом «Export Package to XMI»: дерево модели, " +
			"чтение элемента/пакета/диаграммы, создание и правка элементов и связей, словарь типов ArchiMate 3.2. " +
			"Каждая правка пишется в новый XMI-файл для повторного импорта в EA."},
	{"Публикация в GitHub", "2. Публикация в GitHub",
		"Готово. Release-workflow: пер-ОС бинарники (linux/darwin/windows × amd64/arm64) при бампе версии; " +
			"one-line install.sh / install.ps1; PR-гейт — unit-test и проверка бампа версии."},
	{"Доп функции", "3. Доп функции",
		"В работе. Создание Sparx XML с нуля (ea_new_model / ea_create_root_package / ea_set_root_name, " +
			"ArchiMate3-профиль встроен); глубокое копирование пакетов; создание диаграмм (ea_create_diagram) " +
			"и размещение элементов на них."},
	{"Улучшения отчётов о тестировании", "4. Улучшения отчётов о тестировании",
		"Запланировано. junit / coverage / mutation отчёты и бейджи README на GitHub Pages; сборка нескольких " +
			"XMI в один импортируемый файл (AssembleReport); человекочитаемый отчёт о правках модели."},
	{"Нумерация всех ошибок", "5. Нумерация всех ошибок",
		"Запланировано. Сквозные стабильные коды ошибок (EA-NNNN) во всех сообщениях MCP, логах и документации; " +
			"реестр известных ошибок."},
	{"Отправка bug report", "6. Отправка bug report",
		"Запланировано. Инструмент, которым агент отправляет отчёт об ошибке сопровождающему проекта (номер ошибки, " +
			"контекст вызова, входной файл). Кандидат реализации — создание GitHub Issue."},
	{"Запрос добавления типов", "7. Запрос добавления типов",
		"Запланировано. Функционал для запроса расширения словаря типов: когда сервер встречает неизвестный тип " +
			"при чтении модели, либо когда агенту не хватает текущих типов ArchiMate."},
}

var wpNote = map[string]string{
	"MVP": "client/eaxmi (кодек XMI 2.1), internal/service/sparx (словарь и валидация ArchiMate 3.2), " +
		"internal/mcpserver. Тесты — Cucumber-сценарии (godog).",
	"Публикация в GitHub": ".github/workflows: release (пер-ОС бинарники при бампе mcpserver.Version), " +
		"pr (unit-test + version-check). scripts/install.sh, scripts/install.ps1.",
	"Доп функции": "tools/modelgen; ea_new_model / ea_create_root_package / ea_set_root_name; встраивание " +
		"ArchiMate3-профиля; ea_copy_package; ea_create_diagram; ea_place_on_diagram / ea_move_on_diagram / ea_remove_from_diagram.",
	"Улучшения отчётов о тестировании": "make test-report / test-and-report; report-template/; публикация " +
		"coverage / mutation / junit и бейджей на GitHub Pages; AssembleReport — склейка XMI.",
	"Нумерация всех ошибок": "Единый тип ошибки с кодом EA-NNNN в internal/service/sparx и internal/mcpserver; " +
		"реестр кодов в docs; коды в сообщениях MCP и логах.",
	"Отправка bug report": "Инструмент ea_bug_report: сбор контекста (код ошибки, аргументы, входной файл), " +
		"отправка сопровождающему (GitHub Issue API / вебхук), возврат ссылки агенту.",
	"Запрос добавления типов": "Детект неизвестного стереотипа в client/eaxmi при чтении; инструмент " +
		"ea_request_type (имя, вид, надтип, обоснование); канал запроса сопровождающему; расширение словаря в internal/service/sparx.",
}

func (b *builder) roadmap() {
	root := b.pkg("Sparx EA MCP Server", "План развития")
	wpPkg := b.pkg(root, "Работы")
	reqPkg := b.pkg("Sparx EA MCP Server", "Требования")

	req := b.el(reqPkg, "ArchiMate.Requirement", "Сквозная нумерация ошибок",
		"Каждая ошибка сервера имеет стабильный номер/код (напр. EA-1234), одинаковый в сообщении MCP, логах и "+
			"документации. Код — ключ для поиска в базе известных проблем и для функции отправки bug report.")

	var prev string
	for _, s := range steps {
		plateau := b.el(root, "ArchiMate.Plateau", s.name, s.note)
		wp := b.el(wpPkg, "ArchiMate.WorkPackage", "Работа: "+s.key, wpNote[s.key])
		b.rel("ArchiMate.Realization", wp, plateau, "")
		if prev != "" {
			// The ArchiMate validator only allows Triggering between behaviour
			// elements, so plateau ordering is an Association.
			b.rel("ArchiMate.Association", prev, plateau, "затем")
		}
		prev = plateau
	}

	// Each work package realizes the capability services its step delivers…
	for _, g := range groups {
		if key := deliveredBy[g.name]; key != "" {
			b.rel("ArchiMate.Realization", wpPath+key, svcPath+g.name, "")
		}
	}
	// …and the error-numbering work package realizes the requirement.
	b.rel("ArchiMate.Realization", wpPath+"Нумерация всех ошибок", req, "")
}

// ---------- diagrams ----------

const (
	boxW, boxH = 170, 70
	gapX, gapY = 60, 50
	marginX    = 40
)

func (b *builder) diagrams() {
	b.capabilitiesDiagram()
	b.roadmapDiagram()
}

// capabilitiesDiagram: the component on the left, its functions in a column,
// the matching services in a column to their right.
func (b *builder) capabilitiesDiagram() {
	d := b.dgm("Sparx EA MCP Server/Возможности сервиса", "Обзор возможностей", "Application")

	colFn := marginX + boxW + gapX
	colSvc := colFn + boxW + gapX
	midY := marginX + (len(groups)-1)*(boxH+gapY)/2

	b.place(d, compPath, marginX, midY)
	for i, g := range groups {
		y := marginX + i*(boxH+gapY)
		b.place(d, fnPath+g.name, colFn, y)
		b.place(d, svcPath+g.name, colSvc, y)
	}
}

// roadmapDiagram: the seven plateaus in a row, each work package below it, the
// requirement below its work package.
func (b *builder) roadmapDiagram() {
	d := b.dgm("Sparx EA MCP Server/План развития", "Дорожная карта", "Implementation_Migration")

	for i, s := range steps {
		x := marginX + i*(boxW+gapX)
		b.place(d, plPath+s.name, x, marginX)
		b.place(d, wpPath+s.key, x, marginX+boxH+gapY)
		if s.key == "Нумерация всех ошибок" {
			b.place(d, "Sparx EA MCP Server/Требования/Сквозная нумерация ошибок", x, marginX+2*(boxH+gapY))
		}
	}
}
