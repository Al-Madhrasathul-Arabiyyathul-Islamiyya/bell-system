using Microsoft.UI.Windowing;
using WinRT.Interop;

namespace ClientApp;

/// <summary>
/// Provides application-specific behavior to supplement the default Application class.
/// </summary>
public sealed partial class App : Microsoft.UI.Xaml.Application
{
    readonly Composition.AppHost host = new();
    Window? window;

    /// <summary>
    /// Initializes the singleton application object.  This is the first line of authored code
    /// executed, and as such is the logical equivalent of main() or WinMain().
    /// </summary>
    public App() => InitializeComponent();

    /// <summary>
    /// Gets the application service provider.
    /// </summary>
    public IServiceProvider Services => host.Services;

    /// <summary>
    /// Invoked when the application is launched normally by the end user.
    /// </summary>
    /// <param name="args">Details about the launch request and process.</param>
    protected override void OnLaunched(LaunchActivatedEventArgs args)
    {
        _ = args;

        window ??= new();

        window.Content ??= Services.GetRequiredService<Views.ShellPage>();

        if (ShouldLaunchInFullScreen())
        {
            TryEnterFullScreen(window);
        }

        window.Activate();
    }

    static bool ShouldLaunchInFullScreen() =>
        string.Equals(
            Environment.GetEnvironmentVariable("BELL_SYSTEM_FULLSCREEN"),
            bool.TrueString,
            StringComparison.OrdinalIgnoreCase);

    static void TryEnterFullScreen(Window targetWindow)
    {
        IntPtr windowHandle = WindowNative.GetWindowHandle(targetWindow);
        var windowId = Microsoft.UI.Win32Interop.GetWindowIdFromWindow(windowHandle);
        var appWindow = AppWindow.GetFromWindowId(windowId);
        appWindow.SetPresenter(AppWindowPresenterKind.FullScreen);
    }
}
