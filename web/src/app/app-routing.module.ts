import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {DashboardComponent} from './dashboard/dashboard.component';
import {LiveStreamComponent} from './live-stream/live-stream.component';
import {UserRouteAccessGuard} from './guards/user-route-access.guard';
import {LoginComponent} from './login/login.component';
import {LobbyEntryComponent} from './lobby-entry/lobby-entry.component';

const routes: Routes = [
  { path: '', redirectTo: '/dashboard', pathMatch: 'full' },
  { path: 'dashboard', component: DashboardComponent, canActivate: [UserRouteAccessGuard] },
  { path: 'lobby/:spaceId/stream/:streamId', component: LobbyEntryComponent, canActivate: [UserRouteAccessGuard] },
  { path: 'stream/:id', component: LiveStreamComponent, canActivate: [UserRouteAccessGuard] },
  { path: 'login', component: LoginComponent}
];

@NgModule({
  imports: [
    RouterModule.forRoot(routes)
  ],
  exports: [RouterModule]
})
export class AppRoutingModule { }
