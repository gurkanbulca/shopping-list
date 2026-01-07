package shopping.android.ui

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.compose.viewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import shopping.domain.usecase.LoginUseCase

/**
 * Login screen for user authentication.
 */
@Composable
fun LoginScreen(
    onLoginSuccess: () -> Unit,
    modifier: Modifier = Modifier,
    viewModel: LoginViewModel = viewModel { LoginViewModel.create() }
) {
    val state by viewModel.state.collectAsState()

    Column(
        modifier = modifier
            .fillMaxSize()
            .padding(16.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Text(
            text = "Shopping List",
            style = MaterialTheme.typography.headlineLarge,
            modifier = Modifier.padding(bottom = 32.dp)
        )

        OutlinedTextField(
            value = state.email,
            onValueChange = viewModel::onEmailChanged,
            label = { Text("Email") },
            enabled = !state.isLoading,
            keyboardOptions = KeyboardOptions(
                keyboardType = KeyboardType.Email,
                imeAction = ImeAction.Next
            ),
            modifier = Modifier.fillMaxWidth()
        )

        Spacer(modifier = Modifier.height(16.dp))

        OutlinedTextField(
            value = state.password,
            onValueChange = viewModel::onPasswordChanged,
            label = { Text("Password") },
            enabled = !state.isLoading,
            visualTransformation = PasswordVisualTransformation(),
            keyboardOptions = KeyboardOptions(
                keyboardType = KeyboardType.Password,
                imeAction = ImeAction.Done
            ),
            keyboardActions = KeyboardActions(
                onDone = { viewModel.onLoginClicked(onLoginSuccess) }
            ),
            modifier = Modifier.fillMaxWidth()
        )

        Spacer(modifier = Modifier.height(24.dp))

        state.error?.let { error ->
            Text(
                text = error,
                color = MaterialTheme.colorScheme.error,
                style = MaterialTheme.typography.bodyMedium,
                modifier = Modifier.padding(bottom = 16.dp)
            )
        }

        Button(
            onClick = { viewModel.onLoginClicked(onLoginSuccess) },
            enabled = !state.isLoading && state.email.isNotBlank() && state.password.isNotBlank(),
            modifier = Modifier
                .fillMaxWidth()
                .height(48.dp)
        ) {
            if (state.isLoading) {
                CircularProgressIndicator(
                    modifier = Modifier.size(24.dp),
                    color = MaterialTheme.colorScheme.onPrimary
                )
            } else {
                Text("Login")
            }
        }
    }
}

/**
 * ViewModel for the Login screen.
 */
class LoginViewModel(
    private val loginUseCase: LoginUseCase
) : ViewModel() {

    private val _state = MutableStateFlow(LoginState())
    val state: StateFlow<LoginState> = _state.asStateFlow()

    fun onEmailChanged(email: String) {
        _state.value = _state.value.copy(email = email, error = null)
    }

    fun onPasswordChanged(password: String) {
        _state.value = _state.value.copy(password = password, error = null)
    }

    fun onLoginClicked(onSuccess: () -> Unit) {
        val currentState = _state.value
        if (currentState.isLoading) return

        _state.value = currentState.copy(isLoading = true, error = null)

        viewModelScope.launch {
            val result = loginUseCase(
                email = currentState.email,
                password = currentState.password
            )

            result.fold(
                onSuccess = {
                    _state.value = currentState.copy(isLoading = false, error = null)
                    onSuccess()
                },
                onFailure = { error ->
                    _state.value = currentState.copy(
                        isLoading = false,
                        error = error.message ?: "Login failed"
                    )
                }
            )
        }
    }

    companion object {
        fun create(): LoginViewModel {
            val appGraph = shopping.data.di.AppGraph.get()
            val loginUseCase = LoginUseCase(appGraph.authRepository)
            return LoginViewModel(loginUseCase)
        }
    }
}

/**
 * UI state for the Login screen.
 */
data class LoginState(
    val email: String = "",
    val password: String = "",
    val isLoading: Boolean = false,
    val error: String? = null
)
