namespace ClientApp.Application.ViewModels;

/// <summary>
/// Represents the main dashboard surface for schedule and health information.
/// </summary>
public sealed partial class DashboardViewModel : BaseViewModel
{
    /// <summary>
    /// Initializes the dashboard view model.
    /// </summary>
    public DashboardViewModel() => Title = "Dashboard";
}
