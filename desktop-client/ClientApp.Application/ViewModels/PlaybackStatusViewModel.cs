namespace ClientApp.Application.ViewModels;

/// <summary>
/// Represents the current playback status shown in the UI.
/// </summary>
public sealed partial class PlaybackStatusViewModel : BaseViewModel
{
    /// <summary>
    /// Initializes the playback status view model.
    /// </summary>
    public PlaybackStatusViewModel() => Title = "Playback";
}
