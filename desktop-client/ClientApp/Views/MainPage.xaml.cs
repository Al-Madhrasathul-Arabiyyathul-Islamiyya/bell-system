namespace ClientApp.Views;

/// <summary>
/// A simple page that can be used on its own or navigated to within a Frame.
/// </summary>
sealed partial class MainPage : Page
{
    public MainPage()
    {
        ViewModel = ((App)Microsoft.UI.Xaml.Application.Current).Services.GetRequiredService<MainViewModel>();
        InitializeComponent();
    }

    public MainViewModel ViewModel { get; }
}
