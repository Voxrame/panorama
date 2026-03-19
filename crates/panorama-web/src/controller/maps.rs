use axum::Json;
use axum::extract::State;
use serde::Serialize;
use uuid::Uuid;

use crate::domain;
use crate::usecase::maps;
use crate::controller::AppState;

#[derive(Serialize)]
pub struct MapDto {
    pub id: Uuid,
    pub name: String,
    pub kind: String,
}

#[derive(Serialize)]
pub struct ListMapsResponse {
    pub maps: Vec<MapDto>,
}

pub async fn list(state: State<AppState>) -> Json<ListMapsResponse> {
    let request = maps::ListRequest {};
    let maps = state.maps_usecase.list(request).await.unwrap_or_default();

    let map_dtos: Vec<MapDto> = maps.into_iter().map(map_to_dto).collect();

    Json(ListMapsResponse { maps: map_dtos })
}

fn map_to_dto(map: domain::Map) -> MapDto {
    MapDto {
        id: map.id,
        name: map.name,
        kind: map.kind,
    }
}
