namespace ClientApp.Views;

/// <summary>
/// Root navigation shell for the desktop client.
/// </summary>
sealed partial class ShellPage : Page
{
    public ShellPage()
    {
        InitializeComponent();
        Loaded += OnLoaded;
    }

    void OnLoaded(object sender, RoutedEventArgs e)
    {
        _ = sender;
        _ = e;

        if (ContentFrame.Content is null)
        {
            if (NavigationRoot.MenuItems[0] is NavigationViewItem firstItem)
            {
                NavigationRoot.SelectedItem = firstItem;
            }

            _ = ContentFrame.Navigate(typeof(MainPage));
        }
    }

    void NavigationRoot_SelectionChanged(NavigationView sender, NavigationViewSelectionChangedEventArgs args)
    {
        if (args.SelectedItemContainer is not NavigationViewItem selectedItem)
        {
            return;
        }

        var pageType = selectedItem.Tag?.ToString() switch
        {
            "dashboard" => typeof(MainPage),
            "settings" => typeof(SettingsPage),
            _ => null,
        };

        if (pageType is not null && ContentFrame.CurrentSourcePageType != pageType)
        {
            _ = ContentFrame.Navigate(pageType);
        }
    }
}
