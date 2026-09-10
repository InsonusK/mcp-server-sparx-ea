# Пробелы правил связей ArchiMate

> **Источник истины** — [internal/service/sparx/archimate_relationships.csv](../internal/service/sparx/archimate_relationships.csv)
> (`source,target,relation,status`, status = `allow` | `warn` | `deny`).
> Вьюха: [docs/archimate-relationship-matrix.md](../docs/archimate-relationship-matrix.md).
> Здесь — сгруппированный разбор с черновиками правил.

Сервис уже работает трёхзначно: `deny` — `ea_create_relationship` отказывает,
`warn` — создаёт с предупреждением, `allow` — молча; при чтении модели
`warn`/`deny`-связи помечаются. CSV сгенерирован seed-правилами
(`tools/reltable -seed`): `deny` — где нотация точно запрещает, `allow` — где
аспектные правила точно разрешают, `warn` — всё между.

Задача — уточнить `warn`-строки по фактам из Sparx EA: собрать пример →
`File → Export → Package to XMI` → `ea_element` (посмотреть `ea_type` /
стереотип / direction) → поправить строку в CSV (`warn` → `allow` или `deny`) +
сценарий в
[relationship_create.feature](../internal/service/sparx/features/relationship_create.feature).

## Что сейчас разрешает валидатор

| Связь | Правило |
| --- | --- |
| Association | всегда |
| Specialization | только один и тот же тип элемента |
| Composition, Aggregation | **только один и тот же тип** (иначе «not supported yet») |
| Influence | цель — motivation-элемент |
| Realization | цель motivation → ок; источник motivation и цель не motivation → запрет; иначе ок |
| Assignment | источник active-structure; цель behavior / active / passive |
| Serving | источник active или behavior; цель не passive |
| Access | источник behavior; цель passive |
| Triggering, Flow | **оба — behavior** |

Aspect-классы (`elementAspect`): `Plateau` / `WorkPackage` / `Deliverable` /
`ImplementationEvent` / `Gap` → `implementation`;
`Resource` / `Capability` / `CourseOfAction` / `ValueStream` → `strategy`;
`Location` / `Grouping` → `composite`.

---

## A. Triggering / Flow за пределами behavior ↔ behavior

| # | Пара (пример) | Основание ArchiMate 3.2 | Увер. |
| --- | --- | --- | --- |
| A1 | `Plateau` → `Plateau` (Triggering) | §11: временнáя последовательность плато моделируется triggering | высокая |
| A2 | `WorkPackage` → `WorkPackage` (Triggering / Flow) | порядок работ | высокая |
| A3 | `ImplementationEvent` → `WorkPackage` (Triggering) | событие запускает работу | высокая |
| A4 | `CourseOfAction` → `CourseOfAction` (Triggering) | последовательность действий стратегии | средняя |
| A5 | `ValueStream` → `ValueStream` (Triggering / Flow) | этапы потока ценности | средняя |
| A6 | `ApplicationComponent` → `ApplicationComponent` (Flow) | поток данных между активными структурами | средняя |
| A7 | `BusinessRole` → `BusinessProcess` (Triggering) | нет — это Assignment | **не добавлять** |

Предлагаемое правило:

```go
case "Triggering", "Flow":
    if sa == aspectBehavior && ta == aspectBehavior {
        return true, ""
    }
    if sa == aspectImpl && ta == aspectImpl {        // A1–A3
        return true, ""
    }
    if sa == aspectStrategy && ta == aspectStrategy { // A4–A5 (после проверки)
        return true, ""
    }
    return false, rel + " connects two behaviour, implementation or strategy elements"
```

`Flow` между активными структурами (A6) — отдельной строкой `sa == aspectActive && ta == aspectActive`, только после проверки в EA.

## B. Composition / Aggregation между разными типами

Сейчас всё, кроме «тип в тип», отбивается с «not supported yet». В ArchiMate 3.2
эта матрица широкая.

| # | Пара | Основание | Увер. |
| --- | --- | --- | --- |
| B1 | `Grouping` ⇢ любой элемент | группировка элементов любого типа — назначение Grouping | высокая |
| B2 | `Plateau` ⇢ любой core-элемент, `Plateau` ⇢ `Gap` | плато агрегирует элементы, актуальные в этом состоянии | высокая |
| B3 | `Node` ⇢ `Device` / `SystemSoftware` / `Artifact`; `Device` ⇢ `Artifact` | вложенность технологического слоя | высокая |
| B4 | `Location` ⇢ активные структуры / `Equipment` / `Facility` / `Node` | размещение | высокая |
| B5 | `Product` ⇢ `BusinessService` / `ApplicationService` / `Contract` (Aggregation) | продукт объединяет сервисы и контракт | высокая |
| B6 | `ValueStream` ⇢ `Capability` (Aggregation) | за этапом потока ценности стоит способность | средняя |
| B7 | `ApplicationComponent` ⇢ `ApplicationFunction` | нет — это Assignment | **не добавлять** |

Практичный минимум без полной матрицы: разрешить, если

- источник `Grouping`, `Plateau` или `Location` → цель любая; **или**
- источник и цель в одном layer/aspect (`active`↔`active`, `passive`↔`passive`, `strategy`↔`strategy`, `impl`↔`impl`).

## C. Assignment вне active → *

| # | Пара | Основание | Увер. |
| --- | --- | --- | --- |
| C1 | `Resource` → `Capability` (Assignment) | 3.2 прямым текстом: «resources are assigned to capabilities» | высокая |
| C2 | `BusinessActor` / `BusinessRole` → `WorkPackage` (Assignment) | исполнитель работы | средняя |
| C3 | `Capability` → `CourseOfAction` | нет — Realization / Influence | **не добавлять** |

```go
case "Assignment":
    if sa == aspectActive && (ta == aspectBehavior || ta == aspectActive || ta == aspectPassive) {
        return true, ""
    }
    if src == "Resource" && tgt == "Capability" {   // C1
        return true, ""
    }
    ...
```

## D. Оставить как есть

- **Specialization — только один и тот же тип.** Верно по 3.2.
- **Access — behavior → passive.** Верно.
- **Influence — цель motivation, источник любой.** Верно.
- **Realization.** В основном ок; `WorkPackage` / `Deliverable` → `Plateau` уже
  проходит (правило permissive).
- **Serving.** Ок для типовых случаев; стратегические источники (`Capability serves …`)
  — редкие, не трогать без явной потребности.

## Порядок работ

1. Собрать в EA модель со всеми парами из таблиц A–C (по одному примеру на строку).
2. Провести связи через Quick Linker — EA сам покажет, что разрешает.
3. Экспорт в XMI → фикстура `internal/service/sparx/test/testdata/`.
4. Под каждую подтверждённую строку: правило в `relationshipAllowed`, запись в
   `relEA` (`verified: true`), сценарий в `relationship_create.feature`.
5. Бамп `mcpserver.Version`.
