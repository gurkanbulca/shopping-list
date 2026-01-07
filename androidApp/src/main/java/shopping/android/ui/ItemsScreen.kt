package shopping.android.ui

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextDecoration
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.stateIn
import shopping.domain.model.Item
import shopping.domain.model.ListId
import shopping.domain.model.ShoppingList
import shopping.domain.usecase.ObserveItemsUseCase
import shopping.domain.usecase.ObserveListsUseCase

/**
 * Items screen showing all items in a shopping list (read-only for US1).
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ItemsScreen(
    listId: ListId,
    onBack: () -> Unit,
    modifier: Modifier = Modifier,
    viewModel: ItemsViewModel = viewModel(
        key = listId.value
    ) { ItemsViewModel.create(listId) }
) {
    val items by viewModel.items.collectAsState()
    val listName by viewModel.listName.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(listName ?: "Items") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                }
            )
        }
    ) { paddingValues ->
        when {
            items.isEmpty() -> {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(paddingValues),
                    contentAlignment = Alignment.Center
                ) {
                    Text("No items in this list")
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
                    items(items, key = { it.id.value }) { item ->
                        ItemRow(item = item)
                    }
                }
            }
        }
    }
}

@Composable
private fun ItemRow(
    item: Item,
    modifier: Modifier = Modifier
) {
    Card(
        modifier = modifier.fillMaxWidth(),
        colors = if (item.isPending) {
            CardDefaults.cardColors(
                containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.6f)
            )
        } else {
            CardDefaults.cardColors()
        }
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Column(
                modifier = Modifier.weight(1f)
            ) {
                Text(
                    text = item.name,
                    style = MaterialTheme.typography.bodyLarge,
                    textDecoration = if (item.isPurchased) TextDecoration.LineThrough else null,
                    color = if (item.isPurchased) {
                        MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                    } else {
                        MaterialTheme.colorScheme.onSurface
                    }
                )
                
                if (item.isPending) {
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        text = "Pending...",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.secondary
                    )
                }
                
                if (item.priority > 0) {
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        text = "Priority: ${item.priority}",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.primary
                    )
                }
            }

            if (item.isPurchased) {
                Icon(
                    imageVector = androidx.compose.material.icons.Icons.Default.Check,
                    contentDescription = "Purchased",
                    tint = MaterialTheme.colorScheme.primary
                )
            }
        }
    }
}

/**
 * ViewModel for the Items screen.
 */
class ItemsViewModel(
    private val listId: ListId,
    private val observeItemsUseCase: ObserveItemsUseCase,
    private val observeListsUseCase: ObserveListsUseCase
) : ViewModel() {

    val items: StateFlow<List<Item>> = observeItemsUseCase(listId)
        .stateIn(
            scope = viewModelScope,
            started = SharingStarted.WhileSubscribed(5000),
            initialValue = emptyList()
        )

    val listName: StateFlow<String?> = kotlinx.coroutines.flow.flow {
        val list = observeListsUseCase.getList(listId)
        emit(list?.name)
    }.stateIn(
        scope = viewModelScope,
        started = SharingStarted.WhileSubscribed(5000),
        initialValue = null
    )

    companion object {
        fun create(listId: ListId): ItemsViewModel {
            val appGraph = shopping.data.di.AppGraph.get()
            val observeItemsUseCase = ObserveItemsUseCase(appGraph.itemRepository)
            val observeListsUseCase = ObserveListsUseCase(appGraph.listRepository)
            return ItemsViewModel(listId, observeItemsUseCase, observeListsUseCase)
        }
    }
}
