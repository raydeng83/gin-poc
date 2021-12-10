import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MainLayoutComponent } from './main-layout/main-layout.component';
import {SharedNgZorroAntdModule} from "../shared-ng-zorro-antd.module";
import {RouterModule} from "@angular/router";



@NgModule({
  declarations: [
    MainLayoutComponent
  ],
  imports: [
    CommonModule,
    RouterModule,
    SharedNgZorroAntdModule
  ]
})
export class LayoutModule { }
