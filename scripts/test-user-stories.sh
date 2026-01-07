#!/bin/bash

# Shopping List MVP - User Story Testing Script
# Prerequisites: grpcurl, running API server on localhost:50051, PostgreSQL with migrations applied

set -e

API_HOST="${API_HOST:-localhost:50051}"
PLAINTEXT="${PLAINTEXT:--plaintext}"

echo "=========================================="
echo "Shopping List MVP - User Story Tests"
echo "=========================================="
echo "API Host: $API_HOST"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

success() { echo -e "${GREEN}✓ $1${NC}"; }
fail() { echo -e "${RED}✗ $1${NC}"; exit 1; }
info() { echo -e "${YELLOW}→ $1${NC}"; }

# Generate unique test data
TIMESTAMP=$(date +%s)
TEST_EMAIL="testuser${TIMESTAMP}@example.com"
TEST_PASSWORD="TestPass123"
TEST_NAME="Test User ${TIMESTAMP}"

echo "=========================================="
echo "USER STORY 1: User Authentication"
echo "=========================================="
echo ""

# Test 1.1: Register a new user
info "1.1 Registering new user: $TEST_EMAIL"
REGISTER_RESPONSE=$(grpcurl $PLAINTEXT -d "{
  \"email\": \"$TEST_EMAIL\",
  \"password\": \"$TEST_PASSWORD\",
  \"name\": \"$TEST_NAME\"
}" $API_HOST shopping.v1.AuthService/Register 2>&1) || fail "Register failed: $REGISTER_RESPONSE"

ACCESS_TOKEN=$(echo "$REGISTER_RESPONSE" | grep -o '"accessToken": "[^"]*"' | cut -d'"' -f4)
REFRESH_TOKEN=$(echo "$REGISTER_RESPONSE" | grep -o '"refreshToken": "[^"]*"' | cut -d'"' -f4)
USER_ID=$(echo "$REGISTER_RESPONSE" | grep -o '"id": "[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$ACCESS_TOKEN" ]; then
  fail "No access token in response: $REGISTER_RESPONSE"
fi
success "User registered successfully (ID: $USER_ID)"

# Test 1.2: Login with credentials
info "1.2 Logging in with email and password"
LOGIN_RESPONSE=$(grpcurl $PLAINTEXT -d "{
  \"email\": \"$TEST_EMAIL\",
  \"password\": \"$TEST_PASSWORD\"
}" $API_HOST shopping.v1.AuthService/Login 2>&1) || fail "Login failed: $LOGIN_RESPONSE"

LOGIN_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"accessToken": "[^"]*"' | cut -d'"' -f4)
if [ -z "$LOGIN_TOKEN" ]; then
  fail "No access token in login response"
fi
success "Login successful"

# Test 1.3: Get user profile (authenticated)
info "1.3 Getting user profile"
PROFILE_RESPONSE=$(grpcurl $PLAINTEXT -H "authorization: Bearer $ACCESS_TOKEN" \
  -d "{}" $API_HOST shopping.v1.AuthService/GetMe 2>&1) || fail "GetMe failed: $PROFILE_RESPONSE"

PROFILE_EMAIL=$(echo "$PROFILE_RESPONSE" | grep -o '"email": "[^"]*"' | cut -d'"' -f4)
if [ "$PROFILE_EMAIL" != "$TEST_EMAIL" ]; then
  fail "Profile email mismatch"
fi
success "Profile retrieved successfully"

# Test 1.4: Refresh token
info "1.4 Refreshing access token"
REFRESH_RESPONSE=$(grpcurl $PLAINTEXT -d "{
  \"refreshToken\": \"$REFRESH_TOKEN\"
}" $API_HOST shopping.v1.AuthService/RefreshToken 2>&1) || fail "RefreshToken failed: $REFRESH_RESPONSE"

NEW_ACCESS_TOKEN=$(echo "$REFRESH_RESPONSE" | grep -o '"accessToken": "[^"]*"' | cut -d'"' -f4)
if [ -z "$NEW_ACCESS_TOKEN" ]; then
  fail "No new access token in refresh response"
fi
ACCESS_TOKEN="$NEW_ACCESS_TOKEN"
success "Token refreshed successfully"

echo ""
echo -e "${GREEN}USER STORY 1: PASSED${NC}"
echo ""

echo "=========================================="
echo "USER STORY 2: Group Management"
echo "=========================================="
echo ""

# Test 2.1: Create a group
info "2.1 Creating a new group"
GROUP_NAME="Test Group ${TIMESTAMP}"
CREATE_GROUP_RESPONSE=$(grpcurl $PLAINTEXT -H "authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"name\": \"$GROUP_NAME\",
    \"description\": \"A test group for validation\"
  }" $API_HOST shopping.v1.GroupService/CreateGroup 2>&1) || fail "CreateGroup failed: $CREATE_GROUP_RESPONSE"

GROUP_ID=$(echo "$CREATE_GROUP_RESPONSE" | grep -o '"id": "[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$GROUP_ID" ]; then
  fail "No group ID in response"
fi
success "Group created (ID: $GROUP_ID)"

# Test 2.2: List my groups
info "2.2 Listing user's groups"
LIST_GROUPS_RESPONSE=$(grpcurl $PLAINTEXT -H "authorization: Bearer $ACCESS_TOKEN" \
  -d "{\"pagination\": {\"pageSize\": 10}}" $API_HOST shopping.v1.GroupService/ListMyGroups 2>&1) || fail "ListMyGroups failed: $LIST_GROUPS_RESPONSE"

if ! echo "$LIST_GROUPS_RESPONSE" | grep -q "$GROUP_NAME"; then
  fail "Created group not in list"
fi
success "Groups listed successfully"

# Test 2.3: List group members
info "2.3 Listing group members"
LIST_MEMBERS_RESPONSE=$(grpcurl $PLAINTEXT -H "authorization: Bearer $ACCESS_TOKEN" \
  -d "{\"groupId\": \"$GROUP_ID\", \"pagination\": {\"pageSize\": 10}}" $API_HOST shopping.v1.GroupService/ListMembers 2>&1) || fail "ListMembers failed: $LIST_MEMBERS_RESPONSE"

if ! echo "$LIST_MEMBERS_RESPONSE" | grep -q "MEMBER_ROLE_OWNER"; then
  fail "Owner not found in members"
fi
success "Members listed (owner present)"

# Create second user for invitation test
info "2.4 Creating second user for invitation"
TEST_EMAIL2="testuser2_${TIMESTAMP}@example.com"
REGISTER2_RESPONSE=$(grpcurl $PLAINTEXT -d "{
  \"email\": \"$TEST_EMAIL2\",
  \"password\": \"$TEST_PASSWORD\",
  \"name\": \"Test User 2\"
}" $API_HOST shopping.v1.AuthService/Register 2>&1) || fail "Register user 2 failed"

USER2_TOKEN=$(echo "$REGISTER2_RESPONSE" | grep -o '"accessToken": "[^"]*"' | cut -d'"' -f4)
success "Second user created"

# Test 2.5: Invite member
info "2.5 Inviting second user to group"
INVITE_RESPONSE=$(grpcurl $PLAINTEXT -H "authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"groupId\": \"$GROUP_ID\",
    \"email\": \"$TEST_EMAIL2\",
    \"role\": \"MEMBER_ROLE_MEMBER\"
  }" $API_HOST shopping.v1.GroupService/InviteMember 2>&1) || fail "InviteMember failed: $INVITE_RESPONSE"

if ! echo "$INVITE_RESPONSE" | grep -q "MEMBER_STATUS_INVITED"; then
  fail "Invitation status not INVITED"
fi
success "Member invited successfully"

# Test 2.6: Accept invitation
info "2.6 Second user accepting invitation"
ACCEPT_RESPONSE=$(grpcurl $PLAINTEXT -H "authorization: Bearer $USER2_TOKEN" \
  -d "{\"groupId\": \"$GROUP_ID\"}" $API_HOST shopping.v1.GroupService/AcceptInvite 2>&1) || fail "AcceptInvite failed: $ACCEPT_RESPONSE"

if ! echo "$ACCEPT_RESPONSE" | grep -q "MEMBER_STATUS_ACTIVE"; then
  fail "Member status not ACTIVE after accept"
fi
success "Invitation accepted"

echo ""
echo -e "${GREEN}USER STORY 2: PASSED${NC}"
echo ""

echo "=========================================="
echo "USER STORY 3: Shopping Lists & Items"
echo "=========================================="
echo ""

# Test 3.1: Create a shopping list
info "3.1 Creating a shopping list"
LIST_NAME="Grocery List ${TIMESTAMP}"
CREATE_LIST_RESPONSE=$(grpcurl $PLAINTEXT -H "authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"groupId\": \"$GROUP_ID\",
    \"name\": \"$LIST_NAME\",
    \"description\": \"Weekly groceries\"
  }" $API_HOST shopping.v1.ListService/CreateList 2>&1) || fail "CreateList failed: $CREATE_LIST_RESPONSE"

LIST_ID=$(echo "$CREATE_LIST_RESPONSE" | grep -o '"id": "[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$LIST_ID" ]; then
  fail "No list ID in response"
fi
success "List created (ID: $LIST_ID)"

# Test 3.2: Add items to the list
info "3.2 Adding items to the list"

# Add item 1 - Milk (High priority)
ADD_ITEM1_RESPONSE=$(grpcurl $PLAINTEXT -H "authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"listId\": \"$LIST_ID\",
    \"name\": \"Milk\",
    \"priority\": \"ITEM_PRIORITY_HIGH\",
    \"quantity\": \"2 gallons\"
  }" $API_HOST shopping.v1.ListService/AddItem 2>&1) || fail "AddItem (Milk) failed"

ITEM1_ID=$(echo "$ADD_ITEM1_RESPONSE" | grep -o '"id": "[^"]*"' | head -1 | cut -d'"' -f4)
success "Added: Milk (High priority)"

# Add item 2 - Bread (Medium priority)
ADD_ITEM2_RESPONSE=$(grpcurl $PLAINTEXT -H "authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"listId\": \"$LIST_ID\",
    \"name\": \"Bread\",
    \"priority\": \"ITEM_PRIORITY_MEDIUM\",
    \"quantity\": \"1 loaf\"
  }" $API_HOST shopping.v1.ListService/AddItem 2>&1) || fail "AddItem (Bread) failed"

ITEM2_ID=$(echo "$ADD_ITEM2_RESPONSE" | grep -o '"id": "[^"]*"' | head -1 | cut -d'"' -f4)
success "Added: Bread (Medium priority)"

# Add item 3 - Eggs (Urgent priority)
ADD_ITEM3_RESPONSE=$(grpcurl $PLAINTEXT -H "authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"listId\": \"$LIST_ID\",
    \"name\": \"Eggs\",
    \"priority\": \"ITEM_PRIORITY_URGENT\",
    \"quantity\": \"1 dozen\",
    \"notes\": \"Get organic if available\"
  }" $API_HOST shopping.v1.ListService/AddItem 2>&1) || fail "AddItem (Eggs) failed"

ITEM3_ID=$(echo "$ADD_ITEM3_RESPONSE" | grep -o '"id": "[^"]*"' | head -1 | cut -d'"' -f4)
success "Added: Eggs (Urgent priority)"

# Test 3.3: List shopping lists
info "3.3 Listing shopping lists in group"
LIST_LISTS_RESPONSE=$(grpcurl $PLAINTEXT -H "authorization: Bearer $ACCESS_TOKEN" \
  -d "{\"groupId\": \"$GROUP_ID\"}" $API_HOST shopping.v1.ListService/ListLists 2>&1) || fail "ListLists failed"

if ! echo "$LIST_LISTS_RESPONSE" | grep -q "$LIST_NAME"; then
  fail "Created list not found"
fi
success "Lists retrieved"

# Test 3.4: Toggle purchased status
info "3.4 Marking item as purchased"
TOGGLE_RESPONSE=$(grpcurl $PLAINTEXT -H "authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"itemId\": \"$ITEM1_ID\",
    \"isPurchased\": true,
    \"expectedVersion\": 1
  }" $API_HOST shopping.v1.ListService/TogglePurchased 2>&1) || fail "TogglePurchased failed: $TOGGLE_RESPONSE"

if ! echo "$TOGGLE_RESPONSE" | grep -q '"isPurchased": true'; then
  fail "Item not marked as purchased"
fi
success "Milk marked as purchased"

# Test 3.5: Update item
info "3.5 Updating item details"
UPDATE_ITEM_RESPONSE=$(grpcurl $PLAINTEXT -H "authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"itemId\": \"$ITEM2_ID\",
    \"name\": \"Whole Wheat Bread\",
    \"priority\": \"ITEM_PRIORITY_HIGH\",
    \"quantity\": \"2 loaves\",
    \"expectedVersion\": 1
  }" $API_HOST shopping.v1.ListService/UpdateItem 2>&1) || fail "UpdateItem failed: $UPDATE_ITEM_RESPONSE"

if ! echo "$UPDATE_ITEM_RESPONSE" | grep -q "Whole Wheat Bread"; then
  fail "Item name not updated"
fi
success "Item updated (Bread → Whole Wheat Bread)"

# Test 3.6: Reorder items
info "3.6 Reordering items"
REORDER_RESPONSE=$(grpcurl $PLAINTEXT -H "authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"listId\": \"$LIST_ID\",
    \"itemIds\": [\"$ITEM3_ID\", \"$ITEM2_ID\", \"$ITEM1_ID\"]
  }" $API_HOST shopping.v1.ListService/ReorderItems 2>&1) || fail "ReorderItems failed: $REORDER_RESPONSE"

success "Items reordered"

# Test 3.7: Delete item
info "3.7 Deleting an item"
DELETE_RESPONSE=$(grpcurl $PLAINTEXT -H "authorization: Bearer $ACCESS_TOKEN" \
  -d "{\"itemId\": \"$ITEM1_ID\"}" $API_HOST shopping.v1.ListService/DeleteItem 2>&1) || fail "DeleteItem failed"

success "Item deleted (Milk)"

# Test 3.8: Archive list
info "3.8 Archiving the list"
ARCHIVE_RESPONSE=$(grpcurl $PLAINTEXT -H "authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"listId\": \"$LIST_ID\",
    \"archive\": true
  }" $API_HOST shopping.v1.ListService/ArchiveList 2>&1) || fail "ArchiveList failed"

if ! echo "$ARCHIVE_RESPONSE" | grep -q '"isArchived": true'; then
  fail "List not archived"
fi
success "List archived"

echo ""
echo -e "${GREEN}USER STORY 3: PASSED${NC}"
echo ""

echo "=========================================="
echo -e "${GREEN}ALL USER STORIES PASSED!${NC}"
echo "=========================================="
echo ""
echo "Test Summary:"
echo "  - US1 (Authentication): Register, Login, GetMe, RefreshToken"
echo "  - US2 (Groups): CreateGroup, ListMyGroups, InviteMember, AcceptInvite, ListMembers"
echo "  - US3 (Lists/Items): CreateList, AddItem, TogglePurchased, UpdateItem, ReorderItems, DeleteItem, ArchiveList"
echo ""
echo "MVP is ready for demo/deployment!"
