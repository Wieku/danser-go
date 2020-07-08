#version 330

#define borderEnd 99f/512f //79

uniform sampler2DArray tex;
uniform vec4 col_border;
uniform vec4 col_border1;

in vec2 tex_coord;
out vec4 color;
void main()
{
    vec4 in_color = texture(tex, vec3(tex_coord, 0));

	color = tex_coord.x >= borderEnd ? mix(col_border, vec4(in_color.rgb, col_border.a), in_color.a): in_color*col_border1;//mix(col_border1, col_border, smoothstep(45.0/512.0, 60.0/512.0, tex_coord.x));
	//color = tex_coord.x >= borderEnd ? mix(vec4(col_border.rgb, in_color.a), vec4(in_color.rgb, in_color.a), in_color.a): in_color*col_border1;//mix(col_border1, col_border, smoothstep(45.0/512.0, 60.0/512.0, tex_coord.x));
}