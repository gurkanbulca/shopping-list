# Tasks: Shopping List SaaS MVP

**Input**: Design documents from `/specs/001-shopping-list-mvp/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are OPTIONAL per specification - only foundational test infrastructure included.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Mobile + API**: `api/` for backend, `ios/` or `android/` for clients (future)
- Paths follow plan.md structure: `api/internal/domain/`, `api/internal/transport/`, `api/internal/storage/`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Create project structure per implementation plan in api/
- [x] T002 Initialize Go module with go.mod in api/
- [x] T003 [P] Create buf.yaml and buf.gen.yaml in api/proto/
- [x] T004 [P] Setup go.mod with dependencies: google.golang.org/grpc, google.golang.org/protobuf, github.com/bufbuild/buf
- [x] T005 [P] Create api/cmd/server/main.go placeholder
- [x] T006 [P] Create api/internal/config/config.go structure
- [x] T007 [P] Create api/pkg/auth/ directory structure
- [x] T008 [P] Create api/pkg/errors/ directory structure
- [x] T009 [P] Create api/pkg/validation/ directory structure
- [x] T010 [P] Setup .env.example with required environment variables
- [x] T011 [P] Create .gitignore for Go project
- [x] T012 [P] Create README.md with setup instructions

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T013 Setup PostgreSQL database connection in api/internal/storage/postgres/connection.go
- [x] T014 Setup database migrations framework (goose or golang-migrate) in api/migrations/
- [x] T015 Create initial migration for users table in api/migrations/00001_create_users.sql
- [x] T016 Create initial migration for groups table in api/migrations/00002_create_groups.sql
- [x] T017 Create initial migration for group_members table in api/migrations/00003_create_group_members.sql
- [x] T018 Create initial migration for lists table in api/migrations/00004_create_lists.sql
- [x] T019 Create initial migration for items table in api/migrations/00006_create_items.sql
- [x] T020 Create initial migration for categories table in api/migrations/00005_create_categories.sql
- [x] T021 [P] Copy proto files from specs/001-shopping-list-mvp/contracts/ to api/proto/shopping/v1/
- [x] T022 [P] Generate Go code from protobuf using buf generate in api/proto/
- [x] T023 [P] Implement request ID interceptor in api/internal/transport/interceptors/request_id.go
- [x] T024 [P] Implement logging interceptor with zap in api/internal/transport/interceptors/logging.go
- [x] T025 [P] Implement auth interceptor for JWT parsing in api/internal/transport/interceptors/auth.go
- [x] T026 [P] Implement authorization interceptor for membership checks in api/internal/transport/interceptors/authorization.go
- [x] T027 [P] Implement idempotency interceptor in api/internal/transport/interceptors/idempotency.go
- [x] T028 [P] Implement metrics/tracing interceptor with OpenTelemetry in api/internal/transport/interceptors/metrics.go
- [x] T029 [P] Setup zap logger configuration in api/internal/config/config.go
- [x] T030 [P] Setup OpenTelemetry configuration in api/internal/config/config.go
- [x] T031 [P] Implement JWT token generation and validation in api/pkg/auth/jwt.go
- [x] T032 [P] Implement error mapping to gRPC status codes in api/pkg/errors/mapper.go
- [x] T033 [P] Setup gRPC server with interceptors chain in api/cmd/server/main.go
- [x] T034 [P] Implement environment configuration loading (viper/envconfig) in api/internal/config/config.go
- [x] T035 Create repository interfaces in api/internal/storage/repositories/user_repo.go
- [x] T036 Create repository interfaces in api/internal/storage/repositories/group_repo.go
- [x] T037 Create repository interfaces in api/internal/storage/repositories/list_repo.go
- [x] T038 Create repository interfaces in api/internal/storage/repositories/item_repo.go
- [x] T039 Create repository interfaces in api/internal/storage/repositories/category_repo.go
- [x] T040 Create repository interfaces in api/internal/storage/repositories/sync_repo.go
- [x] T041 Setup sqlc configuration (sqlc.yaml) or pgx connection pool in api/internal/storage/postgres/
- [x] T042 Create tests/integration/ directory structure for integration tests
- [x] T043 Create tests/unit/ directory structure for unit tests

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - User Authentication and Session Management (Priority: P1) 🎯 MVP

**Goal**: Users can register, login, view profile, and refresh tokens. Delivers secure user identity and session management.

**Independent Test**: Register a new user, login to receive tokens, view profile, refresh token. All operations work without other features.

### Implementation for User Story 1

- [x] T044 [P] [US1] Implement User domain model in api/internal/domain/auth/user.go
- [x] T045 [US1] Implement password hashing with bcrypt in api/pkg/auth/password.go
- [x] T046 [US1] Implement UserRepository interface methods in api/internal/storage/repositories/user_repo.go
- [x] T047 [US1] Implement UserRepository PostgreSQL implementation in api/internal/storage/postgres/user_repo.go
- [x] T048 [US1] Create SQL queries for user operations in api/internal/storage/postgres/queries/user.sql (if using sqlc)
- [x] T049 [US1] Implement AuthService domain logic in api/internal/domain/auth/service.go
- [x] T050 [US1] Implement Register use case in api/internal/domain/auth/register.go
- [x] T051 [US1] Implement Login use case in api/internal/domain/auth/login.go
- [x] T052 [US1] Implement RefreshToken use case in api/internal/domain/auth/refresh_token.go
- [x] T053 [US1] Implement GetMe use case in api/internal/domain/auth/get_me.go
- [x] T054 [US1] Implement AuthService gRPC handler in api/internal/transport/grpc/auth_handler.go
- [x] T055 [US1] Implement Register RPC handler in api/internal/transport/grpc/auth_handler.go
- [x] T056 [US1] Implement Login RPC handler in api/internal/transport/grpc/auth_handler.go
- [x] T057 [US1] Implement RefreshToken RPC handler in api/internal/transport/grpc/auth_handler.go
- [x] T058 [US1] Implement GetMe RPC handler in api/internal/transport/grpc/auth_handler.go
- [x] T059 [US1] Register AuthService with gRPC server in api/cmd/server/main.go
- [x] T060 [US1] Add input validation for RegisterRequest in api/pkg/validation/auth.go
- [x] T061 [US1] Add input validation for LoginRequest in api/pkg/validation/auth.go
- [x] T062 [US1] Add error handling for duplicate email/phone in Register in api/internal/domain/auth/register.go
- [x] T063 [US1] Add error handling for invalid credentials in Login in api/internal/domain/auth/login.go
- [x] T064 [US1] Add logging for auth operations in api/internal/domain/auth/service.go

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently. Users can register, login, view profile, and refresh tokens.

---

## Phase 4: User Story 2 - Group Creation and Membership Management (Priority: P1)

**Goal**: Users can create groups, invite members, accept invitations, view members, and manage roles. Delivers multi-tenant organization and collaboration foundation.

**Independent Test**: Create a group, invite a user, accept invitation, view members, update roles. All operations work independently (can use test users).

### Implementation for User Story 2

- [x] T065 [P] [US2] Implement Group domain model in api/internal/domain/group/group.go
- [x] T066 [P] [US2] Implement GroupMember domain model in api/internal/domain/group/member.go
- [x] T067 [US2] Implement GroupRepository interface methods in api/internal/storage/repositories/group_repo.go
- [x] T068 [US2] Implement GroupMemberRepository interface methods in api/internal/storage/repositories/group_repo.go
- [x] T069 [US2] Implement GroupRepository PostgreSQL implementation in api/internal/storage/postgres/group_repo.go
- [x] T070 [US2] Create SQL queries for group operations in api/internal/storage/postgres/queries/group.sql (if using sqlc)
- [x] T071 [US2] Implement GroupService domain logic in api/internal/domain/group/service.go
- [x] T072 [US2] Implement CreateGroup use case in api/internal/domain/group/create_group.go
- [x] T073 [US2] Implement ListMyGroups use case in api/internal/domain/group/list_groups.go
- [x] T074 [US2] Implement InviteMember use case in api/internal/domain/group/invite_member.go
- [x] T075 [US2] Implement AcceptInvite use case in api/internal/domain/group/accept_invite.go
- [x] T076 [US2] Implement ListMembers use case in api/internal/domain/group/list_members.go
- [x] T077 [US2] Implement UpdateMemberRole use case in api/internal/domain/group/update_role.go
- [x] T078 [US2] Implement GroupService gRPC handler in api/internal/transport/grpc/group_handler.go
- [x] T079 [US2] Implement CreateGroup RPC handler in api/internal/transport/grpc/group_handler.go
- [x] T080 [US2] Implement ListMyGroups RPC handler with pagination in api/internal/transport/grpc/group_handler.go
- [x] T081 [US2] Implement InviteMember RPC handler in api/internal/transport/grpc/group_handler.go
- [x] T082 [US2] Implement AcceptInvite RPC handler in api/internal/transport/grpc/group_handler.go
- [x] T083 [US2] Implement ListMembers RPC handler with pagination in api/internal/transport/grpc/group_handler.go
- [x] T084 [US2] Implement UpdateMemberRole RPC handler in api/internal/transport/grpc/group_handler.go
- [x] T085 [US2] Register GroupService with gRPC server in api/cmd/server/main.go
- [x] T086 [US2] Implement membership check helper in api/internal/domain/group/membership.go
- [x] T087 [US2] Add authorization checks in group handlers (verify user is member) in api/internal/transport/grpc/group_handler.go
- [x] T088 [US2] Add input validation for group operations in api/pkg/validation/group.go
- [x] T089 [US2] Add error handling for duplicate invitations in api/internal/domain/group/invite_member.go
- [x] T090 [US2] Add error handling for invalid roles in api/internal/domain/group/update_role.go
- [x] T091 [US2] Add logging for group operations in api/internal/domain/group/service.go

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently. Users can authenticate and manage groups.

---

## Phase 5: User Story 3 - Shopping Lists and Items Management (Priority: P1)

**Goal**: Users can create lists, add items, mark items as purchased, reorder items, update item details, and archive lists. Delivers complete shopping list management functionality.

**Independent Test**: Create a list, add items with priorities, mark items purchased, reorder items, update details, archive list. Works independently (can use test group).

### Implementation for User Story 3

- [x] T092 [P] [US3] Implement List domain model in api/internal/domain/list/list.go
- [x] T093 [P] [US3] Implement Item domain model in api/internal/domain/list/item.go
- [x] T094 [US3] Implement ListRepository interface methods in api/internal/storage/repositories/list_repo.go
- [x] T095 [US3] Implement ItemRepository interface methods in api/internal/storage/repositories/item_repo.go
- [x] T096 [US3] Implement ListRepository PostgreSQL implementation in api/internal/storage/postgres/list_repo.go
- [x] T097 [US3] Implement ItemRepository PostgreSQL implementation in api/internal/storage/postgres/item_repo.go
- [x] T098 [US3] Create SQL queries for list operations in api/internal/storage/postgres/queries/list.sql (if using sqlc)
- [x] T099 [US3] Create SQL queries for item operations in api/internal/storage/postgres/queries/item.sql (if using sqlc)
- [x] T100 [US3] Implement default item sorting query (unpurchased first, priority desc, sort_order asc) in api/internal/storage/postgres/queries/item.sql
- [x] T101 [US3] Implement ListService domain logic in api/internal/domain/list/service.go
- [x] T102 [US3] Implement CreateList use case in api/internal/domain/list/create_list.go
- [x] T103 [US3] Implement ListLists use case with pagination in api/internal/domain/list/list_lists.go
- [x] T104 [US3] Implement UpdateList use case with version check in api/internal/domain/list/update_list.go
- [x] T105 [US3] Implement ArchiveList use case in api/internal/domain/list/archive_list.go
- [x] T106 [US3] Implement AddItem use case in api/internal/domain/list/add_item.go
- [x] T107 [US3] Implement UpdateItem use case with version check in api/internal/domain/list/update_item.go
- [x] T108 [US3] Implement DeleteItem use case in api/internal/domain/list/delete_item.go
- [x] T109 [US3] Implement TogglePurchased use case with version check in api/internal/domain/list/toggle_purchased.go
- [x] T110 [US3] Implement ReorderItems use case (batch update sort_order) in api/internal/domain/list/reorder_items.go
- [x] T111 [US3] Implement version conflict check (FailedPrecondition on mismatch) in api/internal/domain/list/version_check.go
- [x] T112 [US3] Implement ListService gRPC handler in api/internal/transport/grpc/list_handler.go
- [x] T113 [US3] Implement CreateList RPC handler in api/internal/transport/grpc/list_handler.go
- [x] T114 [US3] Implement ListLists RPC handler with pagination in api/internal/transport/grpc/list_handler.go
- [x] T115 [US3] Implement UpdateList RPC handler in api/internal/transport/grpc/list_handler.go
- [x] T116 [US3] Implement ArchiveList RPC handler in api/internal/transport/grpc/list_handler.go
- [x] T117 [US3] Implement AddItem RPC handler in api/internal/transport/grpc/list_handler.go
- [x] T118 [US3] Implement UpdateItem RPC handler in api/internal/transport/grpc/list_handler.go
- [x] T119 [US3] Implement DeleteItem RPC handler in api/internal/transport/grpc/list_handler.go
- [x] T120 [US3] Implement TogglePurchased RPC handler in api/internal/transport/grpc/list_handler.go
- [x] T121 [US3] Implement ReorderItems RPC handler in api/internal/transport/grpc/list_handler.go
- [x] T122 [US3] Register ListService with gRPC server in api/cmd/server/main.go
- [x] T123 [US3] Add authorization checks (verify user is group member) in api/internal/transport/grpc/list_handler.go
- [x] T124 [US3] Add input validation for list operations in api/pkg/validation/list.go
- [x] T125 [US3] Add input validation for item operations in api/pkg/validation/item.go
- [x] T126 [US3] Add error handling for version conflicts in api/internal/domain/list/update_item.go
- [x] T127 [US3] Add error handling for invalid list/item IDs in api/internal/domain/list/service.go
- [x] T128 [US3] Add logging for list operations in api/internal/domain/list/service.go

**Checkpoint**: At this point, User Stories 1, 2, AND 3 should all work independently. Users can authenticate, manage groups, and manage shopping lists.

---

## Phase 6: User Story 4 - Category Management (Priority: P2)

**Goal**: Users can create, update, and delete categories within groups. Items can be assigned to categories. Delivers improved organization and filtering capabilities.

**Independent Test**: Create categories, assign items to categories, update category names, delete categories. Works independently (can use test groups and items).

### Implementation for User Story 4

- [x] T129 [P] [US4] Implement Category domain model in api/internal/domain/category/category.go
- [x] T130 [US4] Implement CategoryRepository interface methods in api/internal/storage/repositories/category_repo.go
- [x] T131 [US4] Implement CategoryRepository PostgreSQL implementation in api/internal/storage/postgres/category_repo.go
- [x] T132 [US4] Create SQL queries for category operations in api/internal/storage/postgres/queries/category.sql (if using sqlc)
- [x] T133 [US4] Implement unique category name constraint per group in api/internal/storage/postgres/queries/category.sql
- [x] T134 [US4] Implement CategoryService domain logic in api/internal/domain/category/service.go
- [x] T135 [US4] Implement UpsertCategory use case in api/internal/domain/category/upsert_category.go
- [x] T136 [US4] Implement ListCategories use case with pagination in api/internal/domain/category/list_categories.go
- [x] T137 [US4] Implement DeleteCategory use case (soft delete - items retain reference) in api/internal/domain/category/delete_category.go
- [x] T138 [US4] Implement CategoryService gRPC handler in api/internal/transport/grpc/category_handler.go
- [x] T139 [US4] Implement UpsertCategory RPC handler in api/internal/transport/grpc/category_handler.go
- [x] T140 [US4] Implement ListCategories RPC handler with pagination in api/internal/transport/grpc/category_handler.go
- [x] T141 [US4] Implement DeleteCategory RPC handler in api/internal/transport/grpc/category_handler.go
- [x] T142 [US4] Register CategoryService with gRPC server in api/cmd/server/main.go
- [x] T143 [US4] Update ItemRepository to support category_id assignment in api/internal/storage/postgres/item_repo.go
- [x] T144 [US4] Add authorization checks (verify user is group member) in api/internal/transport/grpc/category_handler.go
- [x] T145 [US4] Add input validation for category operations in api/pkg/validation/category.go
- [x] T146 [US4] Add error handling for duplicate category names in api/internal/domain/category/upsert_category.go
- [x] T147 [US4] Add logging for category operations in api/internal/domain/category/service.go

**Checkpoint**: At this point, User Stories 1, 2, 3, AND 4 should all work independently. Users can manage categories for better organization.

---

## Phase 7: User Story 5 - Offline Synchronization (Priority: P1)

**Goal**: Users can make changes offline, sync when online, and handle conflicts. Delivers reliable offline-first experience.

**Independent Test**: Make changes offline, sync when online, handle version conflicts. Works independently (can use test data).

### Implementation for User Story 5

- [x] T148 [P] [US5] Create mutation_log table migration (optional) in api/migrations/00007_create_mutation_log.sql
- [x] T149 [US5] Implement SyncRepository interface methods in api/internal/storage/repositories/sync_repo.go
- [x] T150 [US5] Implement SyncRepository PostgreSQL implementation in api/internal/storage/postgres/sync_repo.go
- [x] T151 [US5] Create SQL queries for delta sync (GetDelta) in api/internal/storage/postgres/queries/sync.sql (if using sqlc)
- [x] T152 [US5] Implement cursor-based delta query using sequence or updated_at in api/internal/storage/postgres/queries/sync.sql
- [x] T153 [US5] Implement SyncService domain logic in api/internal/domain/sync/service.go
- [x] T154 [US5] Implement GetDelta use case (pull server changes) in api/internal/domain/sync/get_delta.go
- [x] T155 [US5] Implement PushMutations use case (push client changes) in api/internal/domain/sync/push_mutations.go
- [x] T156 [US5] Implement mutation processing with idempotency check in api/internal/domain/sync/mutation_processor.go
- [x] T157 [US5] Implement version conflict detection in PushMutations in api/internal/domain/sync/push_mutations.go
- [x] T158 [US5] Implement batch mutation handling in api/internal/domain/sync/push_mutations.go
- [x] T159 [US5] Implement SyncService gRPC handler in api/internal/transport/grpc/sync_handler.go
- [x] T160 [US5] Implement GetDelta RPC handler in api/internal/transport/grpc/sync_handler.go
- [x] T161 [US5] Implement PushMutations RPC handler in api/internal/transport/grpc/sync_handler.go
- [x] T162 [US5] Register SyncService with gRPC server in api/cmd/server/main.go
- [x] T163 [US5] Add authorization checks (verify user is group member) in api/internal/transport/grpc/sync_handler.go
- [x] T164 [US5] Add input validation for sync operations in api/pkg/validation/sync.go
- [x] T165 [US5] Add error handling for version conflicts (FailedPrecondition) in api/internal/domain/sync/push_mutations.go
- [x] T166 [US5] Add error handling for invalid mutations in api/internal/domain/sync/mutation_processor.go
- [x] T167 [US5] Add logging for sync operations in api/internal/domain/sync/service.go

**Checkpoint**: At this point, all user stories should work independently. Users can sync offline changes and handle conflicts.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [x] T168 [P] Add rate limiting for Auth RPCs in api/internal/transport/interceptors/rate_limit.go
- [x] T169 [P] Implement PII/token masking in logging interceptor in api/internal/transport/interceptors/logging.go
- [x] T170 [P] Add request ID propagation to all handlers in api/internal/transport/grpc/
- [x] T171 [P] Add OpenTelemetry tracing to all RPC handlers in api/internal/transport/grpc/
- [x] T172 [P] Add metrics collection (request count, error rate, latency p95) in api/internal/transport/interceptors/metrics.go
- [x] T173 [P] Configure gRPC server reflection (dev only) in api/cmd/server/main.go
- [x] T174 [P] Add comprehensive error messages for all error cases in api/pkg/errors/mapper.go
- [x] T175 [P] Add input validation for all RPC requests in api/pkg/validation/
- [x] T176 [P] Add logging for all domain operations in api/internal/domain/
- [x] T177 [P] Update README.md with API usage examples
- [x] T178 [P] Create API documentation from proto files
- [x] T179 [P] Add Dockerfile for containerized deployment
- [x] T180 [P] Add docker-compose.yml for local development
- [x] T181 [P] Add CI/CD configuration for lint, test, migration check
- [x] T182 [P] Run quickstart.md validation and update if needed
- [x] T183 Code cleanup and refactoring across all modules
- [x] T184 Performance optimization (query optimization, connection pooling)
- [x] T185 Security audit (SQL injection prevention, input sanitization)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can proceed sequentially in priority order (US1 → US2 → US3 → US5, then US4)
  - US2 depends on US1 (needs auth)
  - US3 depends on US2 (needs groups)
  - US4 depends on US2 and US3 (needs groups and items)
  - US5 depends on US3 (needs lists/items) and US2 (needs groups)
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Depends on US1 (needs authentication to work)
- **User Story 3 (P1)**: Depends on US2 (needs groups as tenant boundary)
- **User Story 4 (P2)**: Depends on US2 (needs groups) and US3 (needs items)
- **User Story 5 (P1)**: Depends on US3 (needs lists/items) and US2 (needs groups)

### Within Each User Story

- Models before repositories
- Repositories before domain services
- Domain services before gRPC handlers
- Core implementation before validation/error handling
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Models within a story marked [P] can run in parallel
- Repository implementations can be done in parallel (different files)
- gRPC handlers can be implemented in parallel (different files)
- All Polish tasks marked [P] can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all models for User Story 1 together:
Task: "T044 [P] [US1] Implement User domain model in api/internal/domain/auth/user.go"
Task: "T045 [US1] Implement password hashing with bcrypt in api/pkg/auth/password.go"

# Launch repository implementations together:
Task: "T047 [US1] Implement UserRepository PostgreSQL implementation in api/internal/storage/postgres/user_repo.go"
Task: "T048 [US1] Create SQL queries for user operations in api/internal/storage/postgres/queries/user.sql"
```

---

## Implementation Strategy

### MVP First (User Stories 1, 2, 3 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Auth)
4. Complete Phase 4: User Story 2 (Groups)
5. Complete Phase 5: User Story 3 (Lists & Items)
6. **STOP and VALIDATE**: Test User Stories 1, 2, 3 independently
7. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (Auth MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo (Groups!)
4. Add User Story 3 → Test independently → Deploy/Demo (Lists MVP!)
5. Add User Story 5 → Test independently → Deploy/Demo (Sync!)
6. Add User Story 4 → Test independently → Deploy/Demo (Categories!)
7. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (Auth)
   - Developer B: Prepares for User Story 2 (Groups)
3. After US1 completes:
   - Developer A: User Story 2 (Groups)
   - Developer B: Prepares for User Story 3 (Lists)
4. After US2 completes:
   - Developer A: User Story 3 (Lists & Items)
   - Developer B: User Story 5 (Sync) - can start after US3 models ready
5. After US3 completes:
   - Developer A: User Story 4 (Categories)
   - Developer B: Polish & optimization

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
- Version checks are critical for conflict resolution - ensure all update operations include expected_version
- Authorization checks must be in place for all group-scoped operations
- Pagination must be implemented for all list operations

---

## Summary

**Total Tasks**: 185

**Tasks per User Story**:
- Setup: 12 tasks
- Foundational: 31 tasks
- User Story 1 (Auth): 21 tasks
- User Story 2 (Groups): 27 tasks
- User Story 3 (Lists & Items): 37 tasks
- User Story 4 (Categories): 19 tasks
- User Story 5 (Sync): 20 tasks
- Polish: 18 tasks

**Parallel Opportunities**: 85+ tasks marked [P] can run in parallel

**Independent Test Criteria**:
- US1: Register, login, view profile, refresh token
- US2: Create group, invite, accept, view members, update roles
- US3: Create list, add items, mark purchased, reorder, update, archive
- US4: Create categories, assign items, update names, delete
- US5: Offline changes, sync, conflict handling

**Suggested MVP Scope**: User Stories 1, 2, 3 (Auth + Groups + Lists/Items) - 79 tasks excluding setup/foundational

**Format Validation**: ✅ All tasks follow checklist format: `- [ ] [TaskID] [P?] [Story?] Description with file path`
