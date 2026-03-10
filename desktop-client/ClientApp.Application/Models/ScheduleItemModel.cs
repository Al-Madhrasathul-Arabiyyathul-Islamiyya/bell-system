namespace ClientApp.Application.Models;

/// <summary>
/// Represents a bell schedule item prepared for the desktop client.
/// </summary>
/// <param name="Id">Unique schedule item identifier.</param>
/// <param name="Name">Display name of the schedule item.</param>
/// <param name="Time">Scheduled local time.</param>
/// <param name="SessionName">Associated session name, if any.</param>
public sealed record ScheduleItemModel(Guid Id, string Name, TimeOnly Time, string? SessionName);
