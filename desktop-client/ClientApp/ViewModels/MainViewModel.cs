namespace ClientApp.ViewModels;

sealed partial class MainViewModel : BaseViewModel
{
    int count;

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
