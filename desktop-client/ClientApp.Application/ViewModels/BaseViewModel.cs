namespace ClientApp.Application.ViewModels;

/// <summary>
/// Provides common observable state shared by application view models.
/// </summary>
public abstract partial class BaseViewModel : ObservableObject
{
    [ObservableProperty]
    public partial string Title { get; set; } = string.Empty;
}
