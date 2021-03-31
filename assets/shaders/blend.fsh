#version 330
precision highp float;

uniform sampler2DArray tex;

uniform int layers;
uniform int head;

uniform float weights[512];

in vec2 tex_coord;
out vec4 color;

void main()
{
    color = vec4(vec3(0), 1);

    for (int i = layers - 1; i >= 0; i--) {
        color.rgb += texture(tex, vec3(tex_coord, (i+1+head)%layers)).rgb * weights[i];
    }
}