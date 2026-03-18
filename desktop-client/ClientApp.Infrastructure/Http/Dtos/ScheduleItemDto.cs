namespace ClientApp.Infrastructure.Http.Dtos;

/// <summary>
/// Represents a schedule item payload returned by the backend API.
/// </summary>
public sealed class ScheduleItemDto
{
    /// <summary>
    /// Gets or sets the schedule item identifier.
    /// </summary>
    public Guid Id { get; set; }

    /// <summary>
    /// Gets or sets the schedule item name.
    /// </summary>
    public string Name { get; set; } = string.Empty;
}
