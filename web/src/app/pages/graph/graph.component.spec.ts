/**
 * Copyright 2024 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { TestBed } from '@angular/core/testing';
import { GraphComponent } from './graph.component';
import {
  WINDOW_CONNECTION_PROVIDER,
  WindowConnectorService,
} from '../../services/frame-connection/window-connector.service';
import { InMemoryWindowConnectionProvider } from '../../services/frame-connection/window-connection-provider.service';
import { GraphPageDataSource } from 'src/app/services/frame-connection/frames/graph-page-datasource.service';
import { BACKEND_API } from 'src/app/services/api/backend-api-interface';
import { GetConfigResponse } from 'src/app/common/schema/api-types';
import { of } from 'rxjs';

describe('GraphComponent', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      providers: [
        WindowConnectorService,
        {
          provide: WINDOW_CONNECTION_PROVIDER,
          useValue: new InMemoryWindowConnectionProvider(),
        },
        {
          provide: BACKEND_API,
          useValue: {
            getConfig: () => {
              return of<GetConfigResponse>({
                viewerMode: false,
              });
            },
          },
        },
        GraphPageDataSource,
      ],
    }).compileComponents();
  });

  it('should create the app', () => {
    const fixture = TestBed.createComponent(GraphComponent);
    const app = fixture.componentInstance;
    expect(app).toBeTruthy();
  });
});
