namespace ClientApp.Application.Models;

/// <summary>
/// Represents the authenticated user and role driving UI permissions.
/// </summary>
/// <param name="Username">Signed-in username.</param>
/// <param name="Role">Assigned system role.</param>
/// <param name="AssignedSession">Assigned schedule session, if any.</param>
public sealed record UserSessionModel(string Username, string Role, string? AssignedSession);
