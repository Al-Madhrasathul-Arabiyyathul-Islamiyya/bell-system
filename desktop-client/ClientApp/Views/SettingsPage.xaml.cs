namespace ClientApp.Views;

/// <summary>
/// Settings reference page using a Windows 11 style grouped layout.
/// </summary>
sealed partial class SettingsPage : Page
{
    public SettingsPage()
    {
        ViewModel = ((App)Microsoft.UI.Xaml.Application.Current).Services.GetRequiredService<SettingsViewModel>();
        InitializeComponent();
    }

    public SettingsViewModel ViewModel { get; }
}
