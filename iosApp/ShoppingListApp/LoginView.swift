import SwiftUI
import shared

@available(iOS 15.0, *)
struct LoginView: View {
    @StateObject private var viewModel = LoginViewModel()
    @Binding var isAuthenticated: Bool
    
    var body: some View {
        VStack(spacing: 24) {
            Text("Shopping List")
                .font(.largeTitle)
                .fontWeight(.bold)
            
            VStack(spacing: 16) {
                TextField("Email", text: $viewModel.email)
                    .textFieldStyle(.roundedBorder)
                    .textContentType(.emailAddress)
                    .autocapitalization(.none)
                    .keyboardType(.emailAddress)
                    .disabled(viewModel.isLoading)
                
                SecureField("Password", text: $viewModel.password)
                    .textFieldStyle(.roundedBorder)
                    .textContentType(.password)
                    .disabled(viewModel.isLoading)
                
                if let error = viewModel.error {
                    Text(error)
                        .foregroundColor(.red)
                        .font(.caption)
                }
                
                Button(action: {
                    Task {
                        await viewModel.login()
                        if viewModel.error == nil {
                            isAuthenticated = true
                        }
                    }
                }) {
                    if viewModel.isLoading {
                        ProgressView()
                            .progressViewStyle(.circular)
                            .tint(.white)
                    } else {
                        Text("Login")
                    }
                }
                .frame(maxWidth: .infinity)
                .frame(height: 48)
                .background(viewModel.canLogin ? Color.blue : Color.gray)
                .foregroundColor(.white)
                .cornerRadius(8)
                .disabled(!viewModel.canLogin || viewModel.isLoading)
            }
            .padding(.horizontal)
        }
        .padding()
    }
}

@available(iOS 15.0, *)
@MainActor
final class LoginViewModel: ObservableObject {
    @Published var email = ""
    @Published var password = ""
    @Published var isLoading = false
    @Published var error: String?
    
    private lazy var loginUseCase: LoginUseCase = {
        let appGraph = AppDependencies.shared.getAppGraph()
        return LoginUseCase(authRepository: appGraph.authRepository)
    }()
    
    var canLogin: Bool {
        !email.isEmpty && !password.isEmpty
    }
    
    func login() async {
        guard canLogin else { return }
        
        isLoading = true
        error = nil
        
        do {
            let result = try await loginUseCase.invoke(email: email, password: password)
            
            // Check if result is success
            if result.isSuccess() {
                isLoading = false
                error = nil
            } else {
                let errorObj = result.exceptionOrNull()
                isLoading = false
                error = errorObj?.message ?? "Login failed"
            }
        } catch {
            isLoading = false
            self.error = error.localizedDescription
        }
    }
}

@available(iOS 15.0, *)
#Preview {
    LoginView(isAuthenticated: .constant(false))
}
