package shopping.android.ui

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import shopping.domain.model.GroupId
import shopping.domain.model.ListId
import shopping.domain.model.ShoppingList
import shopping.domain.usecase.ObserveListsUseCase
import shopping.domain.usecase.SelectGroupUseCase

/**
 * Lists screen showing all shopping lists in a group.
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ListsScreen(
    groupId: GroupId,
    onListSelected: (ListId) -> Unit,
    onBack: () -> Unit,
    modifier: Modifier = Modifier,
    viewModel: ListsViewModel = viewModel(
        key = groupId.value
    ) { ListsViewModel.create(groupId) }
) {
    val lists by viewModel.lists.collectAsState()
    val groupName by viewModel.groupName.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(groupName ?: "Lists") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                }
            )
        }
    ) { paddingValues ->
        when {
            lists.isEmpty() -> {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(paddingValues),
                    contentAlignment = Alignment.Center
                ) {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.Center
                    ) {
                        CircularProgressIndicator()
                        Spacer(modifier = Modifier.height(16.dp))
                        Text("Loading lists...")
                    }
                }
            }
            else -> {
                LazyColumn(
                    modifier = modifier
                        .fillMaxSize()
                        .padding(paddingValues),
                    contentPadding = PaddingValues(16.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    items(lists, key = { it.id.value }) { list ->
                        ListItem(
                            list = list,
                            onClick = { onListSelected(list.id) }
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun ListItem(
    list: ShoppingList,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Card(
        modifier = modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        ) {
            Text(
                text = list.name,
                style = MaterialTheme.typography.titleMedium
            )
        }
    }
}

/**
 * ViewModel for the Lists screen.
 */
class ListsViewModel(
    private val groupId: GroupId,
    private val observeListsUseCase: ObserveListsUseCase,
    private val selectGroupUseCase: SelectGroupUseCase
) : ViewModel() {

    val lists: StateFlow<List<ShoppingList>> = observeListsUseCase(groupId)
        .stateIn(
            scope = viewModelScope,
            started = SharingStarted.WhileSubscribed(5000),
            initialValue = emptyList()
        )

    val groupName: StateFlow<String?> = selectGroupUseCase.selectedGroup
        .stateIn(
            scope = viewModelScope,
            started = SharingStarted.WhileSubscribed(5000),
            initialValue = null
        )
        .let { flow ->
            kotlinx.coroutines.flow.map(flow) { it?.name }
                .stateIn(
                    scope = viewModelScope,
                    started = SharingStarted.WhileSubscribed(5000),
                    initialValue = null
                )
        }

    init {
        // Select the group when this screen is opened
        viewModelScope.launch {
            selectGroupUseCase(groupId)
        }
    }

    companion object {
        fun create(groupId: GroupId): ListsViewModel {
            val appGraph = shopping.data.di.AppGraph.get()
            val observeListsUseCase = ObserveListsUseCase(appGraph.listRepository)
            val selectGroupUseCase = SelectGroupUseCase(
                appGraph.groupRepository,
                appGraph.syncRepository
            )
            return ListsViewModel(groupId, observeListsUseCase, selectGroupUseCase)
        }
    }
}
