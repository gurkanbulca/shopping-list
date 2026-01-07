# Tasks: Mobile Offline MVP (Android + iOS)

**Input**: Design documents from `/specs/002-shopping-list-mvp/`  
**Prerequisites**: `specs/002-shopping-list-mvp/plan.md`, `specs/002-shopping-list-mvp/spec.md`, `specs/002-shopping-list-mvp/research.md`  
**Tests**: Not required by spec (focus on working MVP + smoke test flows)

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Introduce a KMP + native app workspace and doc scaffolding without breaking the existing Go API repo.

- [X] T001 Create Gradle KMP workspace at repo root (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/settings.gradle.kts`, `/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/build.gradle.kts`)
- [X] T002 Add Gradle wrapper (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/gradlew`, `/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/gradlew.bat`, `/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/gradle/wrapper/gradle-wrapper.properties`)
- [X] T003 [P] Create KMP module skeletons (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/domain/build.gradle.kts`, `/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/data/build.gradle.kts`, `/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/db/build.gradle.kts`, `/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/sync/build.gradle.kts`, `/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/platform/build.gradle.kts`)
- [X] T004 [P] Create Android app module skeleton (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/androidApp/build.gradle.kts`, `/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/androidApp/src/main/AndroidManifest.xml`)
- [X] T005 [P] Create iOS app skeleton and shared framework integration stub (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/iosApp/ShoppingListApp.xcodeproj`, `/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/iosApp/ShoppingListApp/ContentView.swift`)
- [X] T006 [P] Add mobile developer docs entrypoints (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/docs/client-architecture.md`, `/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/docs/sync-state-machine.md`)
- [X] T007 Add feature quickstart placeholder for mobile (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/specs/002-shopping-list-mvp/quickstart.md`)
- [X] T008 Add feature data-model placeholder for mobile local DB (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/specs/002-shopping-list-mvp/data-model.md`)
- [X] T009 Add feature contracts bundle by referencing/copying canonical protos (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/specs/002-shopping-list-mvp/contracts/README.md`)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared core primitives required by all user stories (auth/session, local DB, sync plumbing, multi-tenant safety, and observability hooks).

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [X] T010 Define shared domain models + enums (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/domain/src/commonMain/kotlin/shopping/domain/model/Models.kt`)
- [X] T011 Define repository/usecase interfaces (Auth/Group/List/Item/Category/Sync) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/domain/src/commonMain/kotlin/shopping/domain/repo/Repositories.kt`)
- [X] T012 Define remote data source interfaces (domain-in, proto-out) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/data/src/commonMain/kotlin/shopping/data/remote/RemoteDataSources.kt`)
- [X] T013 Define platform hooks (uuid/time/logger/secure storage) via expect/actual (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/platform/src/commonMain/kotlin/shopping/platform/Platform.kt`)
- [X] T014 [P] Implement Android actuals for platform hooks (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/platform/src/androidMain/kotlin/shopping/platform/Platform.android.kt`)
- [X] T015 [P] Implement iOS actuals for platform hooks (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/platform/src/iosMain/kotlin/shopping/platform/Platform.ios.kt`)
- [X] T016 Design SQLDelight schema (groups, lists, items, categories, mutation_queue, sync_state) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/db/src/commonMain/sqldelight/shopping/db/AppDatabase.sq`)
- [X] T017 Add initial SQLDelight migration file (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/db/src/commonMain/sqldelight/shopping/db/migrations/1.sqm`)
- [X] T018 Implement DB access layer wrappers/DAOs (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/db/src/commonMain/kotlin/shopping/db/DbAccess.kt`)
- [X] T019 Implement data mappers (db ↔ domain) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/data/src/commonMain/kotlin/shopping/data/mapper/DbMappers.kt`)
- [X] T020 Implement sync delta applier (idempotent apply by (entity_type, entity_id, version)) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/sync/src/commonMain/kotlin/shopping/sync/DeltaApplier.kt`)
- [X] T021 Implement sync state store (cursor per group + last_success_sync_at) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/sync/src/commonMain/kotlin/shopping/sync/SyncStateStore.kt`)
- [X] T022 Implement mutation queue store (append/claim/mark-done, batch_id for reorder) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/sync/src/commonMain/kotlin/shopping/sync/MutationQueueStore.kt`)
- [X] T023 Implement sync engine state machine (push queued batch → pull delta loop → update cursor) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/sync/src/commonMain/kotlin/shopping/sync/SyncEngine.kt`)
- [X] T024 Implement repositories/usecases over DB + sync (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/data/src/commonMain/kotlin/shopping/data/repo/RepositoriesImpl.kt`)
- [X] T025 Add shared observable state conventions (Flow-based outputs) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/domain/src/commonMain/kotlin/shopping/domain/state/State.kt`)
- [X] T026 [P] Add Android gRPC adapter boundary package + metadata interceptors skeleton (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/androidApp/src/main/java/shopping/android/grpc/GrpcAdapters.kt`)
- [X] T027 [P] Add iOS gRPC adapter boundary package + metadata injection skeleton (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/iosApp/ShoppingListApp/GrpcAdapters.swift`)
- [X] T028 Implement auth token storage (secure) interface in shared + platform implementations (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/platform/src/commonMain/kotlin/shopping/platform/SecureTokenStore.kt`)
- [X] T029 [P] Implement Android secure token storage (EncryptedSharedPreferences/Keystore) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/androidApp/src/main/java/shopping/android/auth/AndroidSecureTokenStore.kt`)
- [X] T030 [P] Implement iOS secure token storage (Keychain) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/iosApp/ShoppingListApp/IosSecureTokenStore.swift`)
- [X] T031 Implement auth refresh policy in remote adapters (401/Unauthenticated → RefreshToken → bounded retry) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/data/src/commonMain/kotlin/shopping/data/remote/AuthRetryingAdapter.kt`)
- [X] T032 Document sync state machine (1 page) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/docs/sync-state-machine.md`)
- [X] T033 Document client architecture (adapter boundary + offline-first rule) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/docs/client-architecture.md`)

**Checkpoint**: Foundation ready — user story implementation can now begin in parallel.

---

## Phase 3: User Story 1 — Sign in and browse (Priority: P1) 🎯 MVP

**Goal**: Login, list groups, list lists, show items for a list from local DB populated by sync delta.

**Independent Test**: Fresh install → login → groups load → pick group → lists load → pick list → items render and are sorted correctly.

- [X] T034 [P] [US1] Implement AuthRemoteDataSource using AuthService RPCs (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/androidApp/src/main/java/shopping/android/grpc/AuthRemoteDataSourceAndroid.kt`)
- [X] T035 [P] [US1] Implement AuthRemoteDataSource for iOS (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/iosApp/ShoppingListApp/AuthRemoteDataSourceIos.swift`)
- [X] T036 [P] [US1] Implement GroupRemoteDataSource using GroupService RPCs (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/androidApp/src/main/java/shopping/android/grpc/GroupRemoteDataSourceAndroid.kt`)
- [X] T037 [P] [US1] Implement GroupRemoteDataSource for iOS (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/iosApp/ShoppingListApp/GroupRemoteDataSourceIos.swift`)
- [X] T038 [P] [US1] Implement ListRemoteDataSource (ListLists) for Android (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/androidApp/src/main/java/shopping/android/grpc/ListRemoteDataSourceAndroid.kt`)
- [X] T039 [P] [US1] Implement ListRemoteDataSource for iOS (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/iosApp/ShoppingListApp/ListRemoteDataSourceIos.swift`)
- [X] T040 [P] [US1] Implement SyncRemoteDataSource (GetDelta) for Android (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/androidApp/src/main/java/shopping/android/grpc/SyncRemoteDataSourceAndroid.kt`)
- [X] T041 [P] [US1] Implement SyncRemoteDataSource (GetDelta) for iOS (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/iosApp/ShoppingListApp/SyncRemoteDataSourceIos.swift`)
- [X] T042 [US1] Wire shared repositories with platform adapters (manual DI factories) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/data/src/commonMain/kotlin/shopping/data/di/AppGraph.kt`)
- [X] T043 [US1] Implement login usecase (store tokens, load user profile) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/domain/src/commonMain/kotlin/shopping/domain/usecase/LoginUseCase.kt`)
- [X] T044 [US1] Implement group listing usecase (remote → DB upsert → Flow) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/domain/src/commonMain/kotlin/shopping/domain/usecase/ObserveGroupsUseCase.kt`)
- [X] T045 [US1] Implement lists listing usecase (remote → DB upsert → Flow) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/domain/src/commonMain/kotlin/shopping/domain/usecase/ObserveListsUseCase.kt`)
- [X] T046 [US1] Implement "select group" behavior and clear cross-group UI state (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/domain/src/commonMain/kotlin/shopping/domain/usecase/SelectGroupUseCase.kt`)
- [X] T047 [US1] Implement initial full delta sync for selected group (GetDelta loop until has_more=false) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/sync/src/commonMain/kotlin/shopping/sync/FullSync.kt`)
- [X] T048 [US1] Implement observe items usecase (read from DB only, apply sorting rules) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/domain/src/commonMain/kotlin/shopping/domain/usecase/ObserveItemsUseCase.kt`)
- [X] T049 [US1] Handle PermissionDenied by clearing group-scoped local data + surfacing error (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/data/src/commonMain/kotlin/shopping/data/policy/TenantBoundaryPolicy.kt`)

- [X] T050 [P] [US1] Android Compose: Login screen (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/androidApp/src/main/java/shopping/android/ui/LoginScreen.kt`)
- [X] T051 [P] [US1] Android Compose: Groups screen (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/androidApp/src/main/java/shopping/android/ui/GroupsScreen.kt`)
- [X] T052 [P] [US1] Android Compose: Lists screen (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/androidApp/src/main/java/shopping/android/ui/ListsScreen.kt`)
- [X] T053 [P] [US1] Android Compose: Items screen (read-only list for US1) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/androidApp/src/main/java/shopping/android/ui/ItemsScreen.kt`)
- [X] T054 [US1] Android Navigation wiring (login → groups → lists → items) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/androidApp/src/main/java/shopping/android/ui/NavGraph.kt`)
- [X] T055 [P] [US1] iOS SwiftUI: Login view (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/iosApp/ShoppingListApp/LoginView.swift`)
- [X] T056 [P] [US1] iOS SwiftUI: Groups view (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/iosApp/ShoppingListApp/GroupsView.swift`)
- [X] T057 [P] [US1] iOS SwiftUI: Lists view (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/iosApp/ShoppingListApp/ListsView.swift`)
- [X] T058 [P] [US1] iOS SwiftUI: Items view (read-only list for US1) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/iosApp/ShoppingListApp/ItemsView.swift`)
- [X] T059 [US1] iOS NavigationStack wiring (login → groups → lists → items) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/iosApp/ShoppingListApp/NavRoot.swift`)
- [X] T060 [US1] Bridge shared Flow → SwiftUI observable wrapper (minimal) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/iosApp/ShoppingListApp/FlowBridge.swift`)

**Checkpoint**: US1 works end-to-end and is independently testable.

---

## Phase 4: User Story 2 — Offline-first add + toggle + sync (Priority: P2)

**Goal**: Add item and toggle purchased offline with optimistic UI; queue mutations; sync push confirms and clears pending.

**Independent Test**: Disable network → add item + toggle → pending markers appear → enable network → sync runs → pending clears and state matches server.

- [ ] T061 [US2] Add “pending” fields/state to local schema (e.g., pending flags or pending_by_mutation_id) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/db/src/commonMain/sqldelight/shopping/db/AppDatabase.sq`)
- [ ] T062 [US2] Implement add item usecase (local insert + enqueue CREATE mutation) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/domain/src/commonMain/kotlin/shopping/domain/usecase/AddItemUseCase.kt`)
- [ ] T063 [US2] Implement toggle purchased usecase (local update + enqueue UPDATE mutation with expected_version) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/domain/src/commonMain/kotlin/shopping/domain/usecase/TogglePurchasedUseCase.kt`)
- [ ] T064 [US2] Implement mutation encoding (JSON bytes) for CREATE/UPDATE item mutations (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/sync/src/commonMain/kotlin/shopping/sync/MutationEncoding.kt`)
- [ ] T065 [US2] Implement SyncRemoteDataSource.PushMutations on Android (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/androidApp/src/main/java/shopping/android/grpc/SyncRemoteDataSourceAndroid.kt`)
- [ ] T066 [US2] Implement SyncRemoteDataSource.PushMutations on iOS (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/iosApp/ShoppingListApp/SyncRemoteDataSourceIos.swift`)
- [ ] T067 [US2] Implement push loop: claim queued mutations → PushMutations(batch) → mark succeeded/failed (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/sync/src/commonMain/kotlin/shopping/sync/PushWorker.kt`)
- [ ] T068 [US2] Implement error mapping for MutationResult.error_code → domain errors (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/sync/src/commonMain/kotlin/shopping/sync/MutationErrors.kt`)
- [ ] T069 [US2] Handle version mismatch: on precondition failure → delta refresh → server-wins reconcile → bounded retry (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/sync/src/commonMain/kotlin/shopping/sync/ConflictHandler.kt`)
- [ ] T070 [US2] Ensure optimistic UI sorting stays stable with pending items (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/domain/src/commonMain/kotlin/shopping/domain/usecase/ObserveItemsUseCase.kt`)

- [ ] T071 [US2] Android Compose: add item UI (text input + submit) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/androidApp/src/main/java/shopping/android/ui/ItemsScreen.kt`)
- [ ] T072 [US2] Android Compose: purchased toggle UI + pending indicator (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/androidApp/src/main/java/shopping/android/ui/ItemsScreen.kt`)
- [ ] T073 [US2] iOS SwiftUI: add item UI (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/iosApp/ShoppingListApp/ItemsView.swift`)
- [ ] T074 [US2] iOS SwiftUI: purchased toggle UI + pending indicator (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/iosApp/ShoppingListApp/ItemsView.swift`)
- [ ] T075 [US2] Add manual “Sync now” affordance for MVP debugging (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/domain/src/commonMain/kotlin/shopping/domain/usecase/SyncNowUseCase.kt`)

**Checkpoint**: US2 works offline-first and syncs when online.

---

## Phase 5: User Story 3 — Reorder + conflict recovery (Priority: P3)

**Goal**: Reorder items locally, enqueue a batch of mutations, sync applies it, and conflicts refresh+reconcile+retry.

**Independent Test**: Reorder items → local order persists across navigation/app restart → offline reorder queues pending → online sync confirms → conflict scenario recovers.

- [ ] T076 [US3] Implement reorder usecase (update local sort_order for all items + enqueue batch mutations with shared batch_id) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/domain/src/commonMain/kotlin/shopping/domain/usecase/ReorderItemsUseCase.kt`)
- [ ] T077 [US3] Extend mutation queue to support atomic “batch_id” claim/send (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/sync/src/commonMain/kotlin/shopping/sync/MutationQueueStore.kt`)
- [ ] T078 [US3] Encode reorder mutations as per-item UPDATE mutations (entity_type=\"item\", expected_version set) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/sync/src/commonMain/kotlin/shopping/sync/MutationEncoding.kt`)
- [ ] T079 [US3] Update push worker to send a reorder batch as a single PushMutations request (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/sync/src/commonMain/kotlin/shopping/sync/PushWorker.kt`)
- [ ] T080 [US3] Android Compose: drag-drop reorder UI + call reorder usecase (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/androidApp/src/main/java/shopping/android/ui/ItemsScreen.kt`)
- [ ] T081 [US3] iOS SwiftUI: reorder UI + call reorder usecase (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/iosApp/ShoppingListApp/ItemsView.swift`)
- [ ] T082 [US3] Add conflict test harness path (manual steps + expected results) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/specs/002-shopping-list-mvp/quickstart.md`)

**Checkpoint**: US3 reorder and conflict recovery are demonstrable end-to-end.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Cross-story hardening: observability, UX clarity, docs, and operational readiness.

- [ ] T083 [P] Ensure request metadata is consistent (authorization, x-request-id, idempotency-key) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/androidApp/src/main/java/shopping/android/grpc/GrpcAdapters.kt`)
- [ ] T084 [P] Ensure iOS metadata injection consistency (authorization, x-request-id, idempotency-key) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/iosApp/ShoppingListApp/GrpcAdapters.swift`)
- [ ] T085 Add structured logging with token/PII masking in shared core (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/platform/src/commonMain/kotlin/shopping/platform/Logger.kt`)
- [ ] T086 Add “last sync” status display to Items screen (both platforms) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/shared/domain/src/commonMain/kotlin/shopping/domain/state/SyncStatus.kt`)
- [ ] T087 Update feature quickstart with full smoke test checklist (login→groups→lists→items→add/toggle/reorder→sync) (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/specs/002-shopping-list-mvp/quickstart.md`)
- [ ] T088 Update local DB data-model doc with tables and key indices/constraints (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/specs/002-shopping-list-mvp/data-model.md`)
- [ ] T089 Add build/run instructions for Android+iOS in docs (`/home/hichlich/workspace/github.com/gurkanbulca/shopping-list/docs/client-architecture.md`)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately.
- **Foundational (Phase 2)**: Depends on Setup — blocks all user stories.
- **User Stories (Phase 3–5)**: Depend on Foundational.
  - Implement sequentially (P1 → P2 → P3) for MVP, or in parallel after Phase 2 if staffed.
- **Polish (Phase 6)**: Depends on the stories you intend to ship.

### User Story Dependencies

- **US1 (P1)**: Requires sync delta pull (items are sourced via delta); otherwise independent.
- **US2 (P2)**: Builds on US1 navigation + item list display; adds write queue and push.
- **US3 (P3)**: Builds on US2 mutation queue/push; adds reorder batching + explicit conflict demo steps.

### Parallel Opportunities

- Tasks marked **[P]** can run in parallel (different files, no blocking dependencies).
- Platform adapters (Android/iOS) can be implemented in parallel after shared interfaces exist (T012–T013).

---

## Parallel Example: US1

```text
Run in parallel:
- T034 [US1] Android AuthRemoteDataSource (androidApp/.../AuthRemoteDataSourceAndroid.kt)
- T035 [US1] iOS AuthRemoteDataSource (iosApp/.../AuthRemoteDataSourceIos.swift)
- T040 [US1] Android SyncRemoteDataSource.GetDelta (androidApp/.../SyncRemoteDataSourceAndroid.kt)
- T041 [US1] iOS SyncRemoteDataSource.GetDelta (iosApp/.../SyncRemoteDataSourceIos.swift)
```

---

## Implementation Strategy

### MVP First (US1 only)

1. Complete Phase 1 → Phase 2
2. Complete Phase 3 (US1)
3. Stop and validate US1 smoke flow end-to-end

### Incremental Delivery

1. Add US2 (offline writes + sync push) → validate offline/online flow
2. Add US3 (reorder + conflict recovery) → validate reorder persistence + conflict demo
