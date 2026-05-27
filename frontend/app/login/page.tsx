import LoginForm from "./LoginForm";

export default function LoginPage() {
  return (
    <main className="mx-auto flex min-h-[70vh] w-full max-w-5xl flex-col items-center justify-center px-6">
      <div className="w-full max-w-sm">
        <h1 className="mb-1 text-2xl font-semibold tracking-tight text-gray-900">
          Sign in to SaltCRM
        </h1>
        <p className="mb-6 text-sm text-gray-500">
          Use your agent email and password.
        </p>
        <LoginForm />
      </div>
    </main>
  );
}
