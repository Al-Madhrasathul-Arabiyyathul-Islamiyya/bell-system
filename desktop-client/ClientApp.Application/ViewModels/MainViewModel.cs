namespace ClientApp.Application.ViewModels;

/// <summary>
/// Temporary shell view model for the desktop client landing page.
/// </summary>
public sealed partial class MainViewModel : BaseViewModel
{
    int count;

    /// <summary>
    /// Initializes the main page view model.
    /// </summary>
    public MainViewModel() => Title = "Home";

    [ObservableProperty]
    public partial string CountText { get; set; } = "Current count: 0";

    [RelayCommand]
    void Increment()
    {
        count++;
        CountText = $"Current count: {count}";
    }
}
