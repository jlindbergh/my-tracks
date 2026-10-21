import { useState, type RefObject } from 'react';
import Map, {
  Layer,
  NavigationControl,
  Popup,
  Source,
  type LayerProps,
  type MapLayerMouseEvent,
  type MapRef,
} from 'react-map-gl/maplibre';
import { useConfig } from '../config';
import { defaultSportColor, formatDate, formatDistance, sportColors } from '../format';

// OpenFreeMap: free vector basemap, no API key. Data © OpenStreetMap
// contributors (ODbL); attribution is supplied by the style itself.
const basemapStyle = 'https://tiles.openfreemap.org/styles/liberty';

const lineColor = [
  'match',
  ['get', 'sport'],
  ...Object.entries(sportColors).flat(),
  defaultSportColor,
] as unknown as string;

const lineWidth = ['interpolate', ['linear'], ['zoom'], 5, 1.5, 12, 3, 16, 5] as unknown as number;

const lineLayout = { 'line-cap': 'round', 'line-join': 'round' } as const;

const activitiesLayerId = 'activities-line';

const activitiesLayer: LayerProps = {
  id: activitiesLayerId,
  type: 'line',
  layout: lineLayout,
  paint: { 'line-color': lineColor, 'line-width': lineWidth, 'line-opacity': 0.8 },
};

interface PopupInfo {
  longitude: number;
  latitude: number;
  id: string;
  name: string;
  startedAt: number | null;
  distanceM: number;
}

interface Props {
  mapRef: RefObject<MapRef | null>;
  /** Changes whenever the user's data changes, so tiles are refetched. */
  dataVersion: number;
  selectedId: string | null;
  onSelect: (id: string | null) => void;
  onLoad: () => void;
}

export default function TrackMap({ mapRef, dataVersion, selectedId, onSelect, onLoad }: Props) {
  const { tiles } = useConfig();
  const [popup, setPopup] = useState<PopupInfo | null>(null);
  const [cursor, setCursor] = useState<string>('');

  function handleClick(e: MapLayerMouseEvent) {
    const feature = e.features?.[0];
    if (!feature) {
      setPopup(null);
      onSelect(null);
      return;
    }
    const p = feature.properties;
    setPopup({
      longitude: e.lngLat.lng,
      latitude: e.lngLat.lat,
      id: String(p.id),
      name: String(p.name),
      startedAt: typeof p.started_at === 'number' ? p.started_at : null,
      distanceM: Number(p.distance_m),
    });
    onSelect(String(p.id));
  }

  return (
    <Map
      ref={mapRef}
      initialViewState={{ longitude: 15, latitude: 62, zoom: 4 }}
      mapStyle={basemapStyle}
      onLoad={onLoad}
      interactiveLayerIds={[activitiesLayerId]}
      onClick={handleClick}
      onMouseEnter={() => setCursor('pointer')}
      onMouseLeave={() => setCursor('')}
      cursor={cursor}
    >
      <NavigationControl position="top-right" />
      {/* Keyed on dataVersion so new imports show up without a page reload. */}
      <Source
        key={dataVersion}
        id="activities"
        type="vector"
        // MapLibre needs an absolute URL. Not new URL(): it would escape the {z}/{x}/{y} braces.
        tiles={[`${window.location.origin}${tiles.url}`]}
        minzoom={tiles.minZoom}
        maxzoom={tiles.maxZoom}
      >
        <Layer {...activitiesLayer} source-layer={tiles.sourceLayer} />
        <Layer
          id="activities-selected"
          type="line"
          source-layer={tiles.sourceLayer}
          filter={['==', ['get', 'id'], selectedId ?? '']}
          layout={lineLayout}
          paint={{ 'line-color': '#111', 'line-width': 6, 'line-opacity': 0.9 }}
          beforeId={activitiesLayerId}
        />
      </Source>
      {popup && (
        <Popup
          longitude={popup.longitude}
          latitude={popup.latitude}
          onClose={() => setPopup(null)}
          closeOnClick={false}
        >
          <strong>{popup.name}</strong>
          <div>
            {formatDate(popup.startedAt)} · {formatDistance(popup.distanceM)}
          </div>
        </Popup>
      )}
    </Map>
  );
}
