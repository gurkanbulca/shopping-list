package shopping.android.ui

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ExitToApp
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
import shopping.domain.model.Group
import shopping.domain.model.GroupId
import shopping.domain.usecase.LoginUseCase
import shopping.domain.usecase.ObserveGroupsUseCase

/**
 * Groups screen showing all groups the user belongs to.
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun GroupsScreen(
    onGroupSelected: (GroupId) -> Unit,
    onLogout: () -> Unit,
    modifier: Modifier = Modifier,
    viewModel: GroupsViewModel = viewModel { GroupsViewModel.create() }
) {
    val groups by viewModel.groups.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("My Groups") },
                actions = {
                    IconButton(onClick = { viewModel.onLogoutClicked(onLogout) }) {
                        Icon(Icons.Default.ExitToApp, contentDescription = "Logout")
                    }
                }
            )
        }
    ) { paddingValues ->
        when {
            groups.isEmpty() -> {
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
                        Text("Loading groups...")
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
                    items(groups, key = { it.id.value }) { group ->
                        GroupItem(
                            group = group,
                            onClick = { onGroupSelected(group.id) }
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun GroupItem(
    group: Group,
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
                text = group.name,
                style = MaterialTheme.typography.titleMedium
            )
        }
    }
}

/**
 * ViewModel for the Groups screen.
 */
class GroupsViewModel(
    private val observeGroupsUseCase: ObserveGroupsUseCase,
    private val loginUseCase: LoginUseCase
) : ViewModel() {

    val groups: StateFlow<List<Group>> = observeGroupsUseCase()
        .stateIn(
            scope = viewModelScope,
            started = SharingStarted.WhileSubscribed(5000),
            initialValue = emptyList()
        )

    fun onLogoutClicked(onLogout: () -> Unit) {
        viewModelScope.launch {
            loginUseCase.logout()
            onLogout()
        }
    }

    companion object {
        fun create(): GroupsViewModel {
            val appGraph = shopping.data.di.AppGraph.get()
            val observeGroupsUseCase = ObserveGroupsUseCase(appGraph.groupRepository)
            val loginUseCase = LoginUseCase(appGraph.authRepository)
            return GroupsViewModel(observeGroupsUseCase, loginUseCase)
        }
    }
}
