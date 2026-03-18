namespace ClientApp.Application.ViewModels;

/// <summary>
/// Represents the root shell state for the desktop client.
/// </summary>
public sealed partial class MainShellViewModel : BaseViewModel
{
    /// <summary>
    /// Initializes the shell view model.
    /// </summary>
    public MainShellViewModel() => Title = "Bell System";
}
