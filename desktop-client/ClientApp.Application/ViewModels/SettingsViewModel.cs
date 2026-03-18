namespace ClientApp.Application.ViewModels;

/// <summary>
/// Represents editable client configuration state.
/// </summary>
public sealed partial class SettingsViewModel : BaseViewModel
{
    /// <summary>
    /// Initializes the settings view model.
    /// </summary>
    public SettingsViewModel() => Title = "Settings";

    [ObservableProperty]
    public partial string BackendBaseUrl { get; set; } = "https://bell-system.local/api/";

    [ObservableProperty]
    public partial string WebSocketUrl { get; set; } = "wss://bell-system.local/ws";

    [ObservableProperty]
    public partial string CacheDatabasePath { get; set; } = @"C:\ProgramData\BellSystem\client.db";

    [ObservableProperty]
    public partial string DeviceName { get; set; } = "Front Office Panel";

    [ObservableProperty]
    public partial bool LaunchOnStartup { get; set; } = true;

    [ObservableProperty]
    public partial bool PlayPreviewSound { get; set; } = true;

    [ObservableProperty]
    public partial bool Use24HourClock { get; set; } = true;

    [ObservableProperty]
    public partial string SelectedSession { get; set; } = "Morning";

    [ObservableProperty]
    public partial string SelectedAudioOutput { get; set; } = "Speakers (High Definition Audio)";

    [ObservableProperty]
    public partial string SelectedTheme { get; set; } = "Use system setting";
}
