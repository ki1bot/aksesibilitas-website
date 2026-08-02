import type { components } from "@/lib/api/schema";

export type HealthResponse = components["schemas"]["HealthResponse"];
export type User = components["schemas"]["User"];
export type AuthResponse = components["schemas"]["AuthResponse"];
export type Project = components["schemas"]["Project"];
export type ProjectRequest = components["schemas"]["ProjectRequest"];
export type Scan = components["schemas"]["Scan"];
export type ScanStatus = components["schemas"]["ScanStatus"];
export type Violation = components["schemas"]["Violation"];
export type ViolationImpact = components["schemas"]["ViolationImpact"];
export type ViolationNode = components["schemas"]["ViolationNode"];
export type ViolationDetail = components["schemas"]["ViolationDetail"];
export type ReviewStatus = components["schemas"]["ReviewStatus"];
export type ManualReview = components["schemas"]["ManualReview"];
export type ManualReviewItem = components["schemas"]["ManualReviewItem"];
export type ManualReviewResponse =
  components["schemas"]["ManualReviewResponse"];
export type Report = components["schemas"]["Report"];
export type ReportFormat = components["schemas"]["ReportFormat"];
export type ErrorResponse = components["schemas"]["ErrorResponse"];
