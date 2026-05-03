use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    Json,
};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

use crate::models::AppState;

#[derive(Debug, Serialize)]
pub struct RevenueMetrics {
    pub period: String,
    pub total_revenue: f64,
    pub arpu: f64,
    pub mrr: f64,
    pub arr: f64,
    pub revenue_growth_pct: f64,
    pub by_plan: HashMap<String, f64>,
    pub by_region: HashMap<String, f64>,
}

#[derive(Debug, Serialize)]
pub struct UsageAnalytics {
    pub period: String,
    pub total_data_gb: f64,
    pub total_voice_minutes: f64,
    pub total_sms: i64,
    pub average_data_per_user_gb: f64,
    pub peak_usage_hour: i32,
    pub usage_trend: String,
}

#[derive(Debug, Serialize)]
pub struct BillingAnalytics {
    pub period: String,
    pub total_invoiced: f64,
    pub total_collected: f64,
    pub collection_rate_pct: f64,
    pub outstanding_amount: f64,
    pub average_days_to_pay: f64,
    pub by_payment_method: HashMap<String, f64>,
}

#[derive(Debug, Deserialize)]
pub struct AnalyticsQuery {
    pub period: Option<String>,
    pub start_date: Option<String>,
    pub end_date: Option<String>,
}

/// Get revenue metrics
pub async fn get_revenue_metrics(
    State(_state): State<AppState>,
    Query(query): Query<AnalyticsQuery>,
) -> Json<RevenueMetrics> {
    let period = query.period.unwrap_or_else(|| "monthly".to_string());
    
    let mut by_plan = HashMap::new();
    by_plan.insert("basic".to_string(), 1500000.0);
    by_plan.insert("premium".to_string(), 2000000.0);
    by_plan.insert("enterprise".to_string(), 1000000.0);
    
    let mut by_region = HashMap::new();
    by_region.insert("us".to_string(), 2500000.0);
    by_region.insert("eu".to_string(), 1500000.0);
    by_region.insert("apac".to_string(), 500000.0);
    
    Json(RevenueMetrics {
        period,
        total_revenue: 4500000.0,
        arpu: 30.0,
        mrr: 4500000.0,
        arr: 54000000.0,
        revenue_growth_pct: 12.5,
        by_plan,
        by_region,
    })
}

/// Get usage analytics
pub async fn get_usage_analytics(
    State(_state): State<AppState>,
    Query(query): Query<AnalyticsQuery>,
) -> Json<UsageAnalytics> {
    let period = query.period.unwrap_or_else(|| "monthly".to_string());
    
    Json(UsageAnalytics {
        period,
        total_data_gb: 15000000.0,
        total_voice_minutes: 50000000.0,
        total_sms: 25000000,
        average_data_per_user_gb: 100.0,
        peak_usage_hour: 20, // 8 PM
        usage_trend: "increasing".to_string(),
    })
}

/// Get billing analytics
pub async fn get_billing_analytics(
    State(_state): State<AppState>,
    Query(query): Query<AnalyticsQuery>,
) -> Json<BillingAnalytics> {
    let period = query.period.unwrap_or_else(|| "monthly".to_string());
    
    let mut by_payment_method = HashMap::new();
    by_payment_method.insert("credit_card".to_string(), 3000000.0);
    by_payment_method.insert("bank_transfer".to_string(), 1000000.0);
    by_payment_method.insert("digital_wallet".to_string(), 350000.0);
    
    Json(BillingAnalytics {
        period,
        total_invoiced: 4500000.0,
        total_collected: 4350000.0,
        collection_rate_pct: 96.7,
        outstanding_amount: 150000.0,
        average_days_to_pay: 12.5,
        by_payment_method,
    })
}

#[derive(Debug, Serialize)]
pub struct ChurnRiskScore {
    pub subscriber_id: String,
    pub risk_score: f64,
    pub risk_level: String,
    pub factors: Vec<String>,
    pub recommendations: Vec<String>,
}

/// Get churn risk for a subscriber
pub async fn get_churn_risk(
    State(_state): State<AppState>,
    Path(subscriber_id): Path<String>,
) -> Json<ChurnRiskScore> {
    Json(ChurnRiskScore {
        subscriber_id,
        risk_score: 45.5,
        risk_level: "medium".to_string(),
        factors: vec![
            "Decreased usage over 30 days".to_string(),
            "No recent plan upgrades".to_string(),
            "Support tickets increased".to_string(),
        ],
        recommendations: vec![
            "Offer loyalty discount".to_string(),
            "Proactive customer outreach".to_string(),
            "Personalized plan recommendation".to_string(),
        ],
    })
}

#[derive(Debug, Serialize)]
pub struct PricingRecommendation {
    pub plan_id: String,
    pub current_price: f64,
    pub recommended_price: f64,
    pub price_change_pct: f64,
    pub expected_revenue_impact: f64,
    pub confidence: f64,
    pub reasoning: Vec<String>,
}

/// Get pricing recommendation for a plan
pub async fn get_pricing_recommendation(
    State(_state): State<AppState>,
    Path(plan_id): Path<String>,
) -> Json<PricingRecommendation> {
    Json(PricingRecommendation {
        plan_id,
        current_price: 29.99,
        recommended_price: 32.99,
        price_change_pct: 10.0,
        expected_revenue_impact: 165000.0,
        confidence: 85.0,
        reasoning: vec![
            "Market analysis suggests price increase tolerance".to_string(),
            "Competitor prices are higher".to_string(),
            "Strong value proposition".to_string(),
        ],
    })
}
