package shopping.android.ui

import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.navigation.NavHostController
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import shopping.domain.model.GroupId
import shopping.domain.model.ListId

/**
 * Navigation routes for the app.
 */
sealed class Screen(val route: String) {
    data object Login : Screen("login")
    data object Groups : Screen("groups")
    data object Lists : Screen("lists/{groupId}") {
        fun createRoute(groupId: GroupId) = "lists/${groupId.value}"
    }
    data object Items : Screen("items/{listId}") {
        fun createRoute(listId: ListId) = "items/${listId.value}"
    }
}

/**
 * Main navigation graph for the app.
 */
@Composable
fun AppNavHost(
    modifier: Modifier = Modifier,
    navController: NavHostController = rememberNavController(),
    startDestination: String = Screen.Login.route
) {
    NavHost(
        navController = navController,
        startDestination = startDestination,
        modifier = modifier
    ) {
        // Login screen
        composable(Screen.Login.route) {
            LoginScreen(
                onLoginSuccess = {
                    navController.navigate(Screen.Groups.route) {
                        // Clear login from back stack
                        popUpTo(Screen.Login.route) { inclusive = true }
                    }
                }
            )
        }

        // Groups screen
        composable(Screen.Groups.route) {
            GroupsScreen(
                onGroupSelected = { groupId ->
                    navController.navigate(Screen.Lists.createRoute(groupId))
                },
                onLogout = {
                    navController.navigate(Screen.Login.route) {
                        // Clear entire back stack
                        popUpTo(0) { inclusive = true }
                    }
                }
            )
        }

        // Lists screen
        composable(
            route = Screen.Lists.route,
            arguments = listOf(
                navArgument("groupId") { type = NavType.StringType }
            )
        ) { backStackEntry ->
            val groupIdValue = backStackEntry.arguments?.getString("groupId")
                ?: error("groupId is required")
            val groupId = GroupId(groupIdValue)

            ListsScreen(
                groupId = groupId,
                onListSelected = { listId ->
                    navController.navigate(Screen.Items.createRoute(listId))
                },
                onBack = {
                    navController.popBackStack()
                }
            )
        }

        // Items screen
        composable(
            route = Screen.Items.route,
            arguments = listOf(
                navArgument("listId") { type = NavType.StringType }
            )
        ) { backStackEntry ->
            val listIdValue = backStackEntry.arguments?.getString("listId")
                ?: error("listId is required")
            val listId = ListId(listIdValue)

            ItemsScreen(
                listId = listId,
                onBack = {
                    navController.popBackStack()
                }
            )
        }
    }
}
