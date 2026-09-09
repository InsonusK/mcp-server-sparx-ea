// Command modelgen builds the ArchiMate model of this MCP server itself and
// writes it to docs/mcp-server-sparx-ea.xml, an EA "Import Package from XMI"
// file. It uses internal/service/sparx — the same service the MCP tools call —
// so the model is produced exactly as ea_new_model / ea_create_* would.
//
//	go run ./tools/modelgen            # writes docs/mcp-server-sparx-ea.xml
//	go run ./tools/modelgen -o out.xml
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
)

func main() {
	out := flag.String("o", "docs/mcp-server-sparx-ea.xml", "path to write the model to")
	flag.Parse()

	m, err := sparx.NewModel("Sparx EA MCP Server")
	if err != nil {
		log.Fatal(err)
	}
	b := &builder{m: m}

	b.capabilities()
	b.roadmap()

	if b.err != nil {
		log.Fatal(b.err)
	}
	if err := m.Save(*out); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %s (%d elements, %d relationships)\n", *out, b.elems, b.rels)
}

type builder struct {
	m           *sparx.Service
	err         error
	elems, rels int
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

// ---------- capabilities ----------

type group struct {
	name  string
	note  string
	tools string
}

func (b *builder) capabilities() {
	root := b.pkg("Sparx EA MCP Server", "Возможности сервиса")
	appPkg := b.pkg(root, "Приложение")
	fnPkg := b.pkg(root, "Функции")
	svcPkg := b.pkg(root, "Сервисы")

	comp := b.el(appPkg, "ArchiMate.ApplicationComponent", "mcp-server-sparx-ea",
		"MCP-сервер (stdio) над экспортом Sparx EA в XMI 2.1. Читает и редактирует "+
			"копию модели, отдавая файл для повторного импорта в EA. Чистый Go, без SQL.")

	groups := []group{
		{"Чтение модели", "Навигация по дереву и чтение одного элемента, пакета или диаграммы.",
			"ea_model_tree(file); ea_element(file, ref); ea_package(file, ref); ea_diagram(file, ref)"},
		{"Словарь типов ArchiMate", "Список принимаемых сервером типов элементов и связей ArchiMate 3.2.",
			"ea_archimate_types()"},
		{"Управление элементами", "Создание, переименование, смена документации и перенос, удаление элементов ArchiMate.",
			"ea_create_element(file, output, parent, type, name, note); " +
				"ea_update_element(file, output, ref, name?, note?, parent?); " +
				"ea_delete_element(file, output, ref)"},
		{"Управление связями", "Создание связей ArchiMate с проверкой правил и удаление по id или паре концов.",
			"ea_create_relationship(file, output, source, target, type, name?, note?); " +
				"ea_delete_relationship(file, output, id | source+target)"},
		{"Управление пакетами", "Создание, переименование и перенос, глубокое копирование, удаление пакетов.",
			"ea_create_package(file, output, parent, name); " +
				"ea_update_package(file, output, ref, name?, parent?); " +
				"ea_copy_package(file, output, ref, dest); " +
				"ea_delete_package(file, output, ref, cascade?)"},
		{"Размещение на диаграммах", "Добавление, перемещение и снятие элемента на существующей диаграмме (координаты EA, origin — левый верх).",
			"ea_place_on_diagram(file, output, diagram, element, left, top, right, bottom); " +
				"ea_move_on_diagram(...); ea_remove_from_diagram(file, output, diagram, element)"},
		{"Жизненный цикл модели", "Создание модели с нуля (ArchiMate-профиль встроен), добавление и переименование корневого пакета.",
			"ea_new_model(root, output); ea_create_root_package(file, output, name); " +
				"ea_set_root_name(file, output, name, fresh_identity?)"},
	}

	for _, g := range groups {
		fn := b.el(fnPkg, "ArchiMate.ApplicationFunction", g.name, g.note+"\n\nИнструменты: "+g.tools)
		svc := b.el(svcPkg, "ArchiMate.ApplicationService", g.name, g.note)
		b.rel("ArchiMate.Assignment", comp, fn, "")
		b.rel("ArchiMate.Realization", fn, svc, "")
	}
}

// ---------- roadmap ----------

func (b *builder) roadmap() {
	root := b.pkg("Sparx EA MCP Server", "План развития")
	wpPkg := b.pkg(root, "Работы")

	type step struct {
		name     string
		note     string
		delivers []string // ApplicationService names this step's work package realizes
	}
	steps := []step{
		{"MVP", "Готово. Базовые read/write инструменты над экспортом «Export Package to XMI»: " +
			"дерево модели, элементы, связи, пакеты, размещение на диаграммах, словарь типов.",
			[]string{"Чтение модели", "Словарь типов ArchiMate", "Управление элементами",
				"Управление связями", "Управление пакетами", "Размещение на диаграммах"}},
		{"Публикация в GitHub workflow", "Готово. Release-workflow: пер-ОС бинарники при бампе " +
			"mcpserver.Version; PR-гейт — unit-test и проверка бампа версии.", nil},
		{"Создание Sparx XML", "В работе. ea_new_model / ea_create_root_package / ea_set_root_name; " +
			"ArchiMate3-профиль встроен в модель, созданную с нуля, чтобы EA распознавал стереотипы при импорте.",
			[]string{"Жизненный цикл модели"}},
		{"Отладка отчёта", "Запланировано. AssembleReport — сборка нескольких XMI в один импортируемый файл; " +
			"junit / coverage / mutation отчёты и бейджи README на GitHub Pages.", nil},
		{"Управление status", "Запланировано. Чтение и изменение project-атрибутов элемента " +
			"(status / phase / version): Proposed → Approved → Implemented.", nil},
	}

	var prevPlateau string
	svcPath := func(name string) string {
		return "Sparx EA MCP Server/Возможности сервиса/Сервисы/" + name
	}

	for _, s := range steps {
		plateau := b.el(root, "ArchiMate.Plateau", s.name, s.note)
		wp := b.el(wpPkg, "ArchiMate.WorkPackage", s.name, s.note)
		b.rel("ArchiMate.Realization", wp, plateau, "")
		for _, d := range s.delivers {
			b.rel("ArchiMate.Realization", wp, svcPath(d), "")
		}
		if prevPlateau != "" {
			// The ArchiMate validator only allows Triggering between behaviour
			// elements, so plateau ordering is expressed with Association.
			b.rel("ArchiMate.Association", prevPlateau, plateau, "затем")
		}
		prevPlateau = plateau
	}
}
